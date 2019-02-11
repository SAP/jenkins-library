#!/bin/sh

set -ex

LIBRARY_VERSION_UNDER_TEST=$(git log --format="%H" -n 1)

rm -rf workspace
git clone -b infrastructure-integration-test https://github.com/sap/cloud-s4-sdk-book workspace
cp -f jenkins.yml workspace
cd workspace
echo "@Library(\"piper-lib-os@$LIBRARY_VERSION_UNDER_TEST\") _" | cat - Jenkinsfile > temp && mv temp Jenkinsfile
docker run -v /var/run/docker.sock:/var/run/docker.sock -v ${PWD}:/workspace -v /tmp -e CASC_JENKINS_CONFIG=/workspace/jenkins.yml -e CF_PW -e ERP_PW -e BRANCH_NAME=master ppiper/jenkinsfile-runner
