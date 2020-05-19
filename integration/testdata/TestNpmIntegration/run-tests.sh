#!/usr/bin/env bash

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o piper ../../../
docker build -t npm-integration-test .
docker run -v "$PWD":/test npm-integration-test /run-tests-in-container.sh
