#!/bin/sh

# The purpose of this script is to run the binary inside a test container and to ensure its output is stored for assertions
# This is not very elegant, but I have so far not found a better way to save output of a command run via "docker exec"

"$@" >/tmp/test-log.txt 2>&1
