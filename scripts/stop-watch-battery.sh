#!/bin/bash

if [ -f /tmp/.watch-battery.lock ]; then
  pid=$(cat /tmp/.watch-battery.lock)
  if [ -f "/proc/$pid/cmdline" ]; then
    kill -9 $pid
    exit 0
  else
    echo "The watch is not running"
    exit 1
  fi
fi
