# Laptop battery watcher

## Usage:

```sh
watch-battery [-limit=nnn] [-status-path=/full/path/to/status/file] [-cap-path=/full/path/to/capacity/file] [-notify-cmd=/full/path/to/notify-send/command]
```

_limit_ is an integer number between 0 and 100. It's optional. The default value is 10.

_status-path_ is a string which defines the full path to the file holding the **status** of the battery. The default value is _/sys/class/power_supply/BAT0/status_.

_cap-path_ is a string which defines the full path to the file holding the **capacity** of the battery. The default value is _/sys/class/power_supply/BAT0/capacity_.

_notify-cmd_ is a string which defines the full path to the command **notify-send**.

For example:

```sh
watch-battery -limit=30
```

It will notify when the battery level is equal or less than 30%

## Considerations

This implementation is basically for **Linux**. Using the diferent parameters could make it work on different distros with different paths.

It's not suitable for **Windows**, for sure. Maybe valid for **Mac**.
