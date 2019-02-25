#!/bin/bash -e

source ../prepareTests.sh consumer-test-neo

docker run -v /var/run/docker.sock:/var/run/docker.sock -v "${PWD}":/workspace -v /tmp -e CASC_JENKINS_CONFIG=/workspace/jenkins.yml \
    -e CX_INFRA_IT_CF_USERNAME -e CX_INFRA_IT_CF_PASSWORD -e BRANCH_NAME=consumer-test-neo ppiper/jenkinsfile-runner
