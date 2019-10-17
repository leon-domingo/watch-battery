# Laptop battery watcher

## Usage:

```sh
watch-battery [<limit>]
```

_limit_ is an integer number between 0 and 100. It's optional. The default value is 10.

For example:

```sh
watch-battery 30
```

It will notify when the battery level is equal or less than 30%

## Considerations

I'm using specific folders to get the information about the battery. These might be wrong for different operating systems. Definitely, it will only works on **Linux**, and probably not every distro.

I'm also using a **Linux** specific tool which is _notify-send_ to show the messages.
