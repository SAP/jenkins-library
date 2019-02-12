#!/bin/sh

set -ex

LIBRARY_VERSION_UNDER_TEST=$(git log --format="%H" -n 1)
REPOSITORY_UNDER_TEST=${TRAVIS_REPO_SLUG:-SAP\/jenkins-library}

rm -rf workspace
git clone -b consumer-test https://github.com/sap/cloud-s4-sdk-book workspace
cp -f jenkins.yml workspace
cd workspace
sed -i "" -e "s/__REPO_SLUG__/${REPOSITORY_UNDER_TEST}/g" jenkins.yml
echo "@Library(\"piper-library-os@$LIBRARY_VERSION_UNDER_TEST\") _" | cat - Jenkinsfile > temp && mv temp Jenkinsfile
git commit --all --author="piper-testing-bot <null@null.com>" --message="Set piper lib version for test"
docker run -v /var/run/docker.sock:/var/run/docker.sock -v "${PWD}":/workspace -v /tmp -e CASC_JENKINS_CONFIG=/workspace/jenkins.yml -e PPIPER_INFRA_IT_CF_USERNAME -e PPIPER_INFRA_IT_CF_PASSWORD -e BRANCH_NAME=consumer-test ppiper/jenkinsfile-runner
