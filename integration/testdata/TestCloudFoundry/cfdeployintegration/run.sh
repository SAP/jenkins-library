#!/usr/bin/env bash

# The purpose of this script is to provide a continent way to tinker with the test locally.
# It is not run in CI.

set -x

#pushd ../../../..
#CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags release -o piper
#popd || exit 1
docker run --rm -v "$PWD":/project -u root \
    "$(docker build -q .)" \
    /bin/bash -c "/test.sh"

