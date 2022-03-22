#!/usr/bin/env bash

# Run all test if no arguments are given, run a single test function if it is passed as $1
# For example: `./run-tests.sh TestRegistrySetInNpmrc`

TEST_NAME=$1

pushd ..
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags release -o piper

if [[ "$TEST_NAME" ]]
then
    go test -tags=integration -timeout 25m -run "$TEST_NAME" ./integration/...
else
    go test -tags=integration -timeout 25m ./integration/...
fi

popd || exit
