#!/bin/bash

d=$(dirname "$0")
[ !  -z  "$d"  ] &&  d="$d/"

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

export CLASSPATH_FILE='target/cp.txt'
mvn clean test dependency:build-classpath -Dmdep.outputFile=${CLASSPATH_FILE} > /dev/null 2>&1

# --in: is created by the unit tests. It contains a mapping between the test case (name is
# already adjusted).
# --out: Contains a transformed version. The calls to other pipeline steps are resolved in a
# transitive manner. This allows us to report all Jenkins plugin calls (also the calls which
# are performed by other pipeline steps. E.g.: each step includes basically a call to
# handlePipelineStepErrors. The Plugin calls issues by handlePipelineStepErrors are also
# reported for the step calling that auxiliar step).
groovy  "${d}steps" -in target/trackedCalls.json --out "${CALLS}"

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
