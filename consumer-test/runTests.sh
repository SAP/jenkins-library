#!/bin/sh -e

set -x

# todo build piper and load it in the build

mkdir -p workspace/jenkins-configuration
cp testing-jenkins.yml workspace/jenkins-configuration/
cd workspace

docker run -it --rm -u $(id -u):$(id -g) -v "${PWD}":/cx-server/mount/ ppiper/cx-server-companion:latest init-cx-server

./cx-server start

cd ..

docker run -v //var/run/docker.sock:/var/run/docker.sock -v $(pwd):/workspace \
 -e CASC_JENKINS_CONFIG=/workspace/jenkins.yml -e HOST=$(hostname) -e PPIPER_INFRA_IT_TEST_PROJECT \
 ppiper/jenkinsfile-runner
