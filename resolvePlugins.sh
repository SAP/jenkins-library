#!/bin/bash

mvn clean test
groovy steps.groovy

WS_OUT="$(pwd)/jenkins_workspace"
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

[ -f "${RESULT}" ] && cat "${RESULT}"
