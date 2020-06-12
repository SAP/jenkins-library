#!/usr/bin/env bash

# This test is run in integration_mta_build_test.go
# The purpose of this script is to provide a continent way to tinker with the test locally.
# It is not run in CI.

pushd ../../../..
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags release -o piper
popd || exit 1
docker build -t mta-it .
docker run --rm -v "$PWD":/project -u root \
    --mount type=bind,source="$(pwd)"/../../../../piper,target=/piper mta-it \
    /bin/bash -c "/test.sh"
