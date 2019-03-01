#!/bin/bash

function fail() {
    local message="$1"
    local returnCode=${2:-1}
    echo "[ERROR] ${message}" >&2
    exit "${returnCode}"
}

function notify() {

    local state=${1}
    local description=${2}
    local hash=${3}

    echo "[INFO] Notifying about state \"${state}\" for commit \"${hash}\"."

    curl -X POST \
        --fail \
        --silent \
        --output /dev/null \
        --data "{\"state\": \"${state}\", \"target_url\": \"${TRAVIS_BUILD_WEB_URL}\", \"description\": \"${description}\", \"context\": \"integration-tests\"}" \
        --user "${INTEGRATION_TEST_VOTING_USER}:${INTEGRATION_TEST_VOTING_TOKEN}" \
    "https://api.github.com/repos/SAP/jenkins-library/statuses/${hash}" || fail "Cannot send notification. curl return code: $?"
}

#
# In case the build is performed for a pull request TRAVIS_COMMIT is a merge
# commit between the base branch and the PR branch HEAD. That commit is actually built.
# But for notifying about a build status we need the commit which is currenty
# the HEAD of the PR branch.
#
# In case the build is performed for a simple branch (not associated with a PR)
# In this case there is no merge commit between any base branch and HEAD of a PR branch.
# The commit which we need for notifying about a build status is in this case simply
# TRAVIS_COMMIT itself.
#
COMMIT_HASH_FOR_STATUS_NOTIFICATIONS="${TRAVIS_PULL_REQUEST_SHA}"
[ -z "${COMMIT_HASH_FOR_STATUS_NOTIFICATIONS}" ] && COMMIT_HASH_FOR_STATUS_NOTIFICATIONS="${TRAVIS_COMMIT}"

notify "pending" "Integration tests in progress." "${COMMIT_HASH_FOR_STATUS_NOTIFICATIONS}"

WORKSPACES_ROOT=workspaces
[ -e "${WORKSPACES_ROOT}"  ] && rm -rf ${WORKSPACES_ROOT}

TEST_CASES=$(find testCases -name '*.yml')

# This auxiliar thread is needed in order to produce some output while the
# test are running. Otherwise the job will be canceled after 10 minutes without
# output.
while true; do sleep 10; echo "[INFO] Integration tests still running."; done &
notificationThreadPid=$!

declare -a processes
i=0
for f in ${TEST_CASES}
do
    testCase=$(basename "${f%.*}")
    area=$(dirname "${f#*/}")
    echo "[INFO] Running test case \"${testCase}\" in area \"${area}\"."
    TEST_CASE_ROOT="${WORKSPACES_ROOT}/${area}/${testCase}"
    [ -e "${TEST_CASE_ROOT}" ] && rm -rf "${TEST_CASE_ROOT}"
    mkdir -p "${TEST_CASE_ROOT}" || fail "Cannot create test case root director for test case \"${testCase}\"." 1
    source ./runTest.sh "${testCase}" "${TEST_CASE_ROOT}" &> "${TEST_CASE_ROOT}/log.txt" &
    pid=$!
    processes[$i]="${area}/${testCase}:${pid}"
    echo "[INFO] Test case \"${testCase}\" in area \"${area}\" launched. (PID: \"${pid}\")."
    let i=i+1
done

[ "${i}" == 0 ] && fail "No tests has been executed." 1

#
# wait for the test cases and cat the log
for p in "${processes[@]}"
do
    area=$(dirname "${p%:*}")
    testCase=$(basename "${p%:*}")
    processId="${p#*:}"
    echo "[INFO] Waiting for test case \"${testCase}\" in area \"${area}\" (PID: \"${processId}\")."
    wait "${processId}"
    echo "[INFO] Test case \"${testCase}\" in area \"${area}\" finished (PID: \"${processId}\")."
done

kill -PIPE "${notificationThreadPid}" &>/dev/null

#
# provide the logs
for p in "${processes[@]}"
do
    area=$(dirname "${p%:*}")
    testCase=$(basename "${p%:*}")
    echo "[INFO] === START === Logs for test case \"${testCase}\" ===."
    cat "${TEST_CASE_ROOT}/log.txt"
    echo "[INFO] === END === Logs for test case \"${testCase}\" ===."
done

#
# list test case status
echo "[INFO] Build status:"
failure="false"
for p in "${processes[@]}"
do
    status="UNDEFINED"
    area=$(dirname "${p%:*}")
    testCase=$(basename "${p%:*}")
    TEST_CASE_ROOT="${WORKSPACES_ROOT}/${area}/${testCase}"
    if [ -f "${TEST_CASE_ROOT}/SUCCESS" ]
    then
        status="SUCCESS"
    else
        status="FAILURE"
        failure="true"
    fi
    printf "[INFO] %-30s: %s\n" "${testCase}" "${status}"
done

STATUS_DESCRIPTION="The integration tests failed."
STATUS_STATE="failure"

if [ "${failure}" == "false" ]
then
    STATUS_DESCRIPTION="The integration tests succeeded."
    STATUS_STATE="success"
fi

notify "${STATUS_STATE}" "${STATUS_DESCRIPTION}" "${COMMIT_HASH_FOR_STATUS_NOTIFICATIONS}"

[ "${failure}" != "false" ] && fail "Integration tests failed." 1

echo "[INFO] Integration tests succeeded."
exit 0
