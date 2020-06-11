#!/usr/bin/env bash

pushd ../../../..
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags release -o piper
popd || exit 1
docker run --rm -v "$PWD":/project --mount type=bind,source="$(pwd)"/../../../../piper,target=/piper devxci/mbtci:1.0.14 /bin/bash -c "cd /project; /piper mtaBuild"
