#!/bin/bash

d=$(dirname "$0")
[ !  -z  "$d"  ] &&  d="$d/"

export CLASSPATH_FILE='target/cp.txt'
mvn clean test dependency:build-classpath -Dmdep.outputFile=${CLASSPATH_FILE} > /dev/null 2>&1
groovy  "${d}steps" -in target/trackedCalls.json --out target/performedCalls.json

WS_OUT="$(pwd)/documentation/jenkins_workspace"
WS_IN=/workspace

REL_CALLS=calls.json
REL_RESULT=result.json

CALLS="${WS_OUT}/${REL_CALLS}"
RESULT="${WS_OUT}/${REL_RESULT}"

for f in ${CALLS} ${RESULT}
do
    [ -e "${f}" ] && rm -rf "${f}"
done

cp target/performedCalls.json "${CALLS}"

[ -f "${CALLS}" ] || { echo "File \"${CALLS}\" does not exist." ; exit 1; }

docker run \
    -w "${WS_IN}" \
    --env calls="${WS_IN}/${REL_CALLS}" \
    --env result="${WS_IN}/${REL_RESULT}" \
    -v "${WS_OUT}:${WS_IN}"  \
    ppiper/jenkinsfile-runner \
        -ns \
        -f Jenkinsfile \
        --runWorkspace /workspace

[ -f "${RESULT}" ] || { echo "Result file containing step to plugin mapping not found (${RESULT})."; exit 1;  }

groovy -cp "target/classes:$(cat $CLASSPATH_FILE)" "${d}createDocu" "${@}"
