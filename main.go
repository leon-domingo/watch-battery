package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nightlyone/lockfile"
)

func main() {

	const (
		dischargingStatus   string = "Discharging"
		defaultNotifyCmd    string = "/usr/bin/notify-send"
		defaultLimit        int    = 10
		defaultStatusPath   string = "/sys/class/power_supply/BAT0/status"
		defaultCapacityPath string = "/sys/class/power_supply/BAT0/capacity"
	)

	var notifyCmd string
	flag.StringVar(
		&notifyCmd,
		"notify-cmd",
		defaultNotifyCmd,
		"Full path to the notify command.",
	)

	var limit int
	flag.IntVar(
		&limit,
		"limit",
		defaultLimit,
		"The battery level when the notification will happen.",
	)

	var statusPath string
	flag.StringVar(
		&statusPath,
		"status-path",
		defaultStatusPath,
		"Full path to the file where the status of the battery is stored.",
	)

	var capacityPath string
	flag.StringVar(
		&capacityPath,
		"cap-path",
		defaultCapacityPath,
		"Full path to the file where the current capacity of the battery is stored",
	)

	flag.Parse()

	var err error

	// handle lock file
	lock, err := lockfile.New(filepath.Join(os.TempDir(), ".watch-battery.lock"))
	if err != nil {
		io.WriteString(os.Stderr, "Lock is already acquired\n")
		os.Exit(1)
	}

	// trying to acquire the lock
	err = lock.TryLock()
	if err != nil {
		io.WriteString(os.Stderr, "Error trying to acquire the lock\n")
		os.Exit(1)
	}

	defer lock.Unlock()

	// Usage:
	//   $ watch-battery [-limit=<limit>] [-status-path="/path/to/status/file"] [-cap-path="/path/to/capacity/file"]
	//   <limit> is an integer number between 0 and 100. It's optional. The default value is 10
	//   For example:
	//   $ watch-battery -limit=30
	//   It will notify when the battery level is equal or less than 30%
	if limit < 0 || limit > 100 {
		io.WriteString(os.Stderr, fmt.Sprintf("The given limit (%d) is not correct. Only integer values between 0 and 100 are allowed.\n", limit))
		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Done()

	go func() {
		reachedLimit := false
		charging := false
		n := 0
		m := 0
		for {
			fileStatus, err := os.Open(statusPath)
			if err != nil {
				io.WriteString(os.Stderr, "Error opening file of battery status\n")
				os.Exit(1)
			}

			var dataStatus = make([]byte, 15)
			_, err = fileStatus.Read(dataStatus)
			if err != nil {
				io.WriteString(os.Stderr, "Error reading battery status\n")
				os.Exit(1)
			}
			fileStatus.Close()

			// only on DISCHARGING status
			if strings.Trim(string(dataStatus), "\n\x00 ") == dischargingStatus {

				charging = false

				fileCapacity, err := os.Open(capacityPath)
				if err != nil {
					io.WriteString(os.Stderr, "Error opening file with battery capacity\n")
					os.Exit(1)
				}

				dataCapacity := make([]byte, 5)
				_, err = fileCapacity.Read(dataCapacity)
				if err != nil {
					io.WriteString(os.Stderr, "Error reading battery capacity\n")
					os.Exit(1)
				}
				fileCapacity.Close()

				perc, _ := strconv.Atoi(strings.Trim(string(dataCapacity), "\n\x00 "))
				if perc <= limit {
					reachedLimit = true
					if n%5 == 0 {
						reachedLimit = false
					}

					if !reachedLimit {
						var procAttr os.ProcAttr
						procAttr.Files = []*os.File{
							os.Stdin,
							os.Stdout,
							os.Stderr,
						}

						argv := []string{
							notifyCmd,
							"--urgency=normal",
							"--expire-time=0",
							"--app-name=Battery",
							"--icon=battery-empty",
							fmt.Sprintf("Low battery (%d%%)", perc),
							fmt.Sprintf("the limit of %d%% was reached", limit),
						}

						os.StartProcess(argv[0], argv, &procAttr)
					}

					n++
				}

				now := fmt.Sprintf("%s", time.Now().Format(time.RFC3339))
				io.WriteString(os.Stdout, fmt.Sprintf("%s %d%%\n", now, perc))

				// perc-limit
				if m != 0 && m%10 == 0 {
					io.WriteString(os.Stdout, fmt.Sprintf("...%d%%\n", perc-limit))
				}
			} else {
				if !charging {
					// write just one line to STDOUT
					io.WriteString(os.Stdout, "---Charging---\n")
					charging = true
				}
			}

			time.Sleep(2 * time.Second)
			m++
		}
	}()

	wg.Wait()
}
