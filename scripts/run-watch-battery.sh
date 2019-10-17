#!/bin/bash

# limit = 30%
# stdout > /tmp/watch-battery.log
# stderr > /dev/null

watch-battery -limit=30 1>> /tmp/watch-battery.log 2> /dev/null &
