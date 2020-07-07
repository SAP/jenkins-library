#!/usr/bin/env bash

set -e
set -x

cd /project
mvn package
cf api
