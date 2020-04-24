#!/bin/bash

d=$(dirname "$0")
[ !  -z  "$d"  ] &&  d="$d/"

WS_OUT="$(pwd)/documentation/jenkins_workspace"
WS_IN=/workspace

STEP_CALL_MAPPING_FILE_NAME=step_calls_mapping.json
PLUGIN_MAPPING_FILE_NAME=plugin_mapping.json

CALLS="${WS_OUT}/${STEP_CALL_MAPPING_FILE_NAME}"
PLUGIN_MAPPING="${WS_OUT}/${PLUGIN_MAPPING_FILE_NAME}"

for f in ${CALLS} ${PLUGIN_MAPPING}
do
    [ -e "${f}" ] && rm -rf "${f}"
done

export CLASSPATH_FILE='target/cp.txt'
mvn --batch-mode --show-version clean test dependency:build-classpath -Dmdep.outputFile=${CLASSPATH_FILE}

if [ "$?" != "0" ];then
    echo "[ERROR] maven test / build-classpath failed"
    exit 1
fi

if [ ! -f "${CLASSPATH_FILE}" ];then
    echo "[ERROR] Classpath file required for docu generation does not exist"
    exit 1
fi

# --in: is created by the unit tests. It contains a mapping between the test case (name is
# already adjusted).
# --out: Contains a transformed version. The calls to other pipeline steps are resolved in a
# transitive manner. This allows us to report all Jenkins plugin calls (also the calls which
# are performed by other pipeline steps. E.g.: each step includes basically a call to
# handlePipelineStepErrors. The Plugin calls issues by handlePipelineStepErrors are also
# reported for the step calling that auxiliar step).
groovy  "${d}resolveTransitiveCalls" -in target/trackedCalls.json --out "${CALLS}"

[ -f "${CALLS}" ] || { echo "File \"${CALLS}\" does not exist." ; exit 1; }

docker run \
    -w "${WS_IN}" \
    --env calls="${WS_IN}/${STEP_CALL_MAPPING_FILE_NAME}" \
    --env result="${WS_IN}/${PLUGIN_MAPPING_FILE_NAME}" \
    -v "${WS_OUT}:${WS_IN}"  \
    ppiper/jenkinsfile-runner \
        -ns \
        -f Jenkinsfile \
        --runWorkspace /workspace

[ -f "${PLUGIN_MAPPING}" ] || { echo "Result file containing step to plugin mapping not found (${PLUGIN_MAPPING})."; exit 1;  }

groovy -cp "target/classes:$(cat $CLASSPATH_FILE)" "${d}createDocu" "${@}"
