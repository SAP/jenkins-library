#!/usr/bin/env bash

# Run all test if no arguments are given, run tests if they've passed as arguments
# For example: ./run-tests.sh TestNexusIntegration TestNPMIntegration

pushd ..
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags release -o piper

if [[ "$*" ]]
then
    for testName in "$@"
    do
        go test -v -tags integration -run "$testName" ./integration/...
    done
else
    go test -v -tags integration ./integration/...
fi

popd || exit
