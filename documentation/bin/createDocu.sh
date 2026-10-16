#!/bin/bash

d=$(dirname "$0")
[ !  -z  "$d"  ] &&  d="$d/"

WS_OUT="$(pwd)/documentation/jenkins_workspace"
WS_IN=/workspace

STEP_CALL_MAPPING_FILE_NAME=step_calls_mapping.json
PLUGIN_MAPPING_FILE_NAME=plugin_mapping.json

CALLS="${WS_OUT}/${STEP_CALL_MAPPING_FILE_NAME}"
PLUGIN_MAPPING="${WS_OUT}/${PLUGIN_MAPPING_FILE_NAME}"

JENKINSFILE_RUNNER_IMAGE="${JENKINSFILE_RUNNER_IMAGE:-ppiper/jenkinsfile-runner:latest}"

for f in ${CALLS} ${PLUGIN_MAPPING}
do
    [ -e "${f}" ] && rm -rf "${f}"
done

# The Jenkins runner image is only needed after the (multi-minute) Maven run below,
# so download it in the background and let it overlap with the build. A failure here
# is not fatal: the "docker run" further down still pulls on demand, and a locally
# present image keeps working offline.
echo "[INFO] Pulling ${JENKINSFILE_RUNNER_IMAGE} in the background."
docker pull "${JENKINSFILE_RUNNER_IMAGE}" > /tmp/jenkinsfile-runner-pull.log 2>&1 &
DOCKER_PULL_PID=$!

export CLASSPATH_FILE='target/cp.txt'
TRACKED_CALLS_FILE='target/trackedCalls.json'

# Both inputs below are by-products of the Groovy unit tests: target/trackedCalls.json is
# written by test/groovy/util/StepTracker.groovy, target/cp.txt by dependency:build-classpath.
# In CI the "Groovy" workflow already runs that build once, so it sets PIPER_SKIP_MAVEN and
# hands the results over instead of paying for a second full test run.
if [ "${PIPER_SKIP_MAVEN}" = "true" ];then
    echo "[INFO] PIPER_SKIP_MAVEN is set, reusing the existing Maven output."
else
    # Quality gates are the job of a plain "mvn verify"; this build exists solely to produce
    # the two files above, so the ones that cost real time are switched off:
    #  - jacoco: coverage instrumentation over the whole unit test suite. argLine has to be
    #    re-supplied because jacoco:prepare-agent normally sets it and surefire inherits the
    #    value from the Jenkins plugin parent POM (-Djava.awt.headless=true).
    #  - enforcer: its requirePluginVersions rule resolves metadata for every plugin.
    #  - animal-sniffer: JDK signature check, irrelevant for documentation.
    mvn --batch-mode --show-version clean test dependency:build-classpath \
        -Dmdep.outputFile=${CLASSPATH_FILE} \
        -Djacoco.skip=true \
        -DargLine=-Djava.awt.headless=true \
        -Denforcer.skip=true \
        -Danimal.sniffer.skip=true

    if [ "$?" != "0" ];then
        echo "[ERROR] maven test / build-classpath failed"
        exit 1
    fi
fi

for f in "${CLASSPATH_FILE}" "${TRACKED_CALLS_FILE}"
do
    if [ ! -f "${f}" ];then
        echo "[ERROR] \"${f}\" is required for docu generation but does not exist."
        echo "[ERROR] Run this script without PIPER_SKIP_MAVEN, or make the Maven build hand the file over."
        exit 1
    fi
done

# Run the helper scripts with the project's own Groovy (org.codehaus.groovy:groovy-all is a
# project dependency and therefore part of ${CLASSPATH_FILE}) instead of requiring a separate
# Groovy installation on PATH. "groovy -cp X script" is exactly "java -cp <groovy>:X
# groovy.ui.GroovyMain script" - that is what the groovy launcher script does internally.
GROOVY_CP="target/classes:$(cat $CLASSPATH_FILE)"
groovyRun() {
    java -cp "${GROOVY_CP}" groovy.ui.GroovyMain "$@"
}

# --in: is created by the unit tests. It contains a mapping between the test case (name is
# already adjusted).
# --out: Contains a transformed version. The calls to other pipeline steps are resolved in a
# transitive manner. This allows us to report all Jenkins plugin calls (also the calls which
# are performed by other pipeline steps. E.g.: each step includes basically a call to
# handlePipelineStepErrors. The Plugin calls issues by handlePipelineStepErrors are also
# reported for the step calling that auxiliar step).
groovyRun "${d}resolveTransitiveCalls.groovy" -in "${TRACKED_CALLS_FILE}" --out "${CALLS}"

[ -f "${CALLS}" ] || { echo "File \"${CALLS}\" does not exist." ; exit 1; }

wait "${DOCKER_PULL_PID}" || echo "[WARN] Background pull of ${JENKINSFILE_RUNNER_IMAGE} failed, see /tmp/jenkinsfile-runner-pull.log. Falling back to the image already present locally."

docker run \
    -w "${WS_IN}" \
    --env calls="${WS_IN}/${STEP_CALL_MAPPING_FILE_NAME}" \
    --env result="${WS_IN}/${PLUGIN_MAPPING_FILE_NAME}" \
    -v "${WS_OUT}:${WS_IN}"  \
    "${JENKINSFILE_RUNNER_IMAGE}" \
        -ns \
        -f Jenkinsfile \
        --runWorkspace /workspace

[ -f "${PLUGIN_MAPPING}" ] || { echo "Result file containing step to plugin mapping not found (${PLUGIN_MAPPING})."; exit 1;  }

groovyRun "${d}createDocu.groovy" "${@}"
