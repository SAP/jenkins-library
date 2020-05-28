#!/usr/bin/env bash

pushd ..
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags release -o piper
go test -tags=integration -timeout 20m ./integration/...
popd
