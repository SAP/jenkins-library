#!/bin/bash

WORKSPACES_ROOT=workspaces
LOG_ROOT="logs"
[ -e "${WORKSPACES_ROOT}"  ] && rm -rf ${WORKSPACES_ROOT}
[ -e "${LOG_ROOT}"  ] && rm -rf ${LOG_ROOT}

mkdir -p "${LOG_ROOT}"

for f in `find . -type f -depth 2 -name '*.yml'`
do
  testCase=`basename ${f%.*}`
  area=`dirname ${f#*/}`
  source runTest.sh "${area}" "${testCase}" &> "logs/${area}-${testCase}.log"

done
