#!/bin/bash

WORKSPACES_ROOT=workspaces
[ -e "${WORKSPACES_ROOT}"  ] && rm -rf ${WORKSPACES_ROOT}

for f in `find . -type f -depth 2 -name '*.yml'`
do
  testCase=`basename ${f%.*}`
  area=`dirname ${f#*/}`
  TEST_CASE_ROOT="${WORKSPACES_ROOT}/${area}/${testCase}"
  [ -e "${TEST_CASE_ROOT}" ] && rm -rf "${TEST_CASE_ROOT}"
  mkdir -p "${TEST_CASE_ROOT}"
  source runTest.sh "${area}" "${testCase}" "${TEST_CASE_ROOT}" &> "${TEST_CASE_ROOT}/log.txt"

done
