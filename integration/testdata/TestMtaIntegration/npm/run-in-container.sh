#!/usr/bin/env bash

set -e
set -x

cd /project
/piper mtaBuild

[ -f /project/test-mta-js.mtar ] || (echo "Assertion failed, file test-mta-js.mtar must exist"; exit 1)
