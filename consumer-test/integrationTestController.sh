#!/bin/bash

WORKSPACES_ROOT=workspaces
[ -e "${WORKSPACES_ROOT}"  ] && rm -rf ${WORKSPACES_ROOT}

TEST_CASES=$(find '.' -depth 2 -name '*.yml')

i=0
for f in ${TEST_CASES}
do
    testCase=$(basename "${f%.*}")
    area=$(dirname "${f#*/}")
    echo "[INFO] Running test case \"${testCase}\" in area \"${area}\"."
    TEST_CASE_ROOT="${WORKSPACES_ROOT}/${area}/${testCase}"
    [ -e "${TEST_CASE_ROOT}" ] && rm -rf "${TEST_CASE_ROOT}"
    mkdir -p "${TEST_CASE_ROOT}"
    source ./runTest.sh "${testCase}" "${TEST_CASE_ROOT}" &> "${TEST_CASE_ROOT}/log.txt" &
    pid=$!
    processes[$i]="${testCase}:${pid}"
    echo "[INFO] Test case \"${testCase}\" in area \"${area}\" launched. (PID: \"${pid}\")."
    let i=i+1
done

[ "${i}" == 0 ] && { echo "No tests has been executed."; exit 1;  }

#
# wait for the test cases and cat the log
for p in "${processes[@]}"
do
    testCase=${p%:*}
    processId=${p#*:}
    echo "[INFO] Waiting for test case \"${testCase}\" (PID: \"${processId}\")."
    wait "${processId}"
    echo "[INFO] Test case \"${testCase}\" finished (PID: \"${processId}\")."
    echo "[INFO] <START> Logs for test case \"${testCase}\"."
    cat "${TEST_CASE_ROOT}/log.txt"
    echo "[INFO] <END> Logs for test case \"${testCase}\"."
done

#
# list test case status
echo "[INFO] Build status:"
failure="false"
for p in "${processes[@]}"
do
    status="UNDEFINED"
    testCase=${p%:*}
    if [ -f "${TEST_CASE_ROOT}/SUCCESS" ]
    then
        status="SUCCESS"
    else
        status="FAILURE"
        failure="true"
    fi
    printf "[INFO] %-30s: %s\n" "${testCase}" ${status}
done

if [ "${failure}" == "true" ]
then
    echo "[WARNING] There are test failures. Check earlier log for details."
    exit 1
fi

echo "[INFO] Integration tests succeeded."
exit 0
