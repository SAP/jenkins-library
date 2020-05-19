#!/usr/bin/env bash

# Fail script if any command returns non-zero exit code
set -e

function show_error_log() {
    if [ -f test-log.txt ]; then
        echo Test failed, log:
        cat test-log.txt
    fi
    popd
}
trap show_error_log ERR

cd /test

pushd registrySetInFlags
/test/piper npmExecuteScripts --install --runScripts=ci-build,ci-backend-unit-test --sapNpmRegistry=https://foo.bar >test-log.txt 2>&1
# Expect line starting with the registry url caused by ci-build run-script
grep --quiet "^info  npmExecuteScripts - https://foo.bar" test-log.txt
rm test-log.txt
popd

pushd registrySetInNpmrc
/test/piper npmExecuteScripts --install --runScripts=ci-build,ci-backend-unit-test >test-log.txt 2>&1
# Expect line starting with the registry url caused by ci-build run-script
grep --quiet "^info  npmExecuteScripts - https://example.com" test-log.txt
rm test-log.txt
popd

pushd registryWithTwoModules
/test/piper npmExecuteScripts --install --runScripts=ci-build,ci-backend-unit-test --sapNpmRegistry=https://foo.bar  >test-log.txt 2>&1
grep --quiet "^info  npmExecuteScripts - https://example.com" test-log.txt
grep --quiet "^info  npmExecuteScripts - https://foo.bar" test-log.txt
rm test-log.txt
popd
