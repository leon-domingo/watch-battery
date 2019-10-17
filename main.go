package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nightlyone/lockfile"
)

func main() {

	var err error

	// handle lock file
	lock, err := lockfile.New(filepath.Join(os.TempDir(), ".watch-battery.lock"))
	if err != nil {
		os.Stderr.Write([]byte("Lock is already acquired\n"))
		os.Exit(1)
	}

	// trying to acquire the lock
	err = lock.TryLock()
	if err != nil {
		os.Stderr.Write([]byte("Error trying to acquire the lock\n"))
		os.Exit(1)
	}

	defer lock.Unlock()

	// Usage:
	//   $ watch-battery [<limit>]
	//   <limit> is an integer number between 0 and 100. It's optional. The default value is 10
	//   For example:
	//   $ watch-battery 30
	//   It will notify when the battery level is equal or less than 30%
	limit := 10
	if len(os.Args) > 1 {
		limit, err = strconv.Atoi(os.Args[1])
		if err != nil {
			os.Stderr.Write([]byte(fmt.Sprintf("The given limit (%v) is not correct. Only integer values between 0 and 100 are allowed.\n", os.Args[1])))
			os.Exit(1)
		}

		if limit < 0 || limit > 100 {
			os.Stderr.Write([]byte(fmt.Sprintf("The given limit (%d) is not correct. Only integer values between 0 and 100 are allowed.\n", limit)))
			os.Exit(1)
		}
	} else {
		os.Stdout.Write([]byte(fmt.Sprintf("A default LIMIT of %d%% will be used\n", limit)))
	}

	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Done()

	go func(limit int) {
		reachedLimit := false
		charging := false
		n := 0
		for {
			fileStatus, err := os.Open("/sys/class/power_supply/BAT0/status")
			if err != nil {
				os.Stderr.Write([]byte("Error opening file of battery status\n"))
				os.Exit(1)
			}

			var dataStatus = make([]byte, 15)
			_, err = fileStatus.Read(dataStatus)
			if err != nil {
				os.Stderr.Write([]byte("Error reading battery status\n"))
				os.Exit(1)
			}
			fileStatus.Close()

			// only on DISCHARGING status
			if strings.Trim(string(dataStatus), "\n\x00 ") == "Discharging" {

				charging = false

				fileCapacity, err := os.Open("/sys/class/power_supply/BAT0/capacity")
				if err != nil {
					os.Stderr.Write([]byte("Error opening file with battery capacity\n"))
					os.Exit(1)
				}

				dataCapacity := make([]byte, 5)
				_, err = fileCapacity.Read(dataCapacity)
				if err != nil {
					os.Stderr.Write([]byte("Error reading battery capacity\n"))
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
							"/usr/bin/notify-send",
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

				os.Stdout.Write([]byte(fmt.Sprintf("%d%% (until limit %d%%)\n", perc, perc-limit)))
			} else {
				if !charging {
					// write just one line to STDOUT
					os.Stdout.Write([]byte("---Charging---\n"))
					charging = true
				}
			}

			time.Sleep(2 * time.Second)
		}
	}(limit)

	wg.Wait()
}
