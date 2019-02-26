#!/bin/bash

WORKSPACE=workspace
[ -e "${WORKSPACE}"  ] && rm -rf ${WORKSPACE}

for f in `find . -type f -depth 2 -name '*.yml'`
do
  testCase=`basename ${f%.*}`
  area=`dirname ${f#*/}`
  echo "${area}/${testCase}"
  source runTest.sh ${testCase}
done
