#!/usr/bin/env bash

set -e

cd /test

pushd registrySetInFlags
/test/piper npmExecuteScripts --install --runScripts=ci-build,ci-backend-unit-test --sapNpmRegistry=https://foo.bar  > test-log.txt 2>&1
grep --quiet "Setting environment: npm_config_@sap:registry=https://foo.bar" test-log.txt
rm test-log.txt
popd

pushd registrySetInNpmrc
/test/piper npmExecuteScripts --install --runScripts=ci-build,ci-backend-unit-test  > test-log.txt 2>&1
grep --quiet "Discovered pre-configured npm registry https://example.com" test-log.txt
rm test-log.txt
popd
