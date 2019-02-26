#!/usr/bin/env bash

TEST_CASE=$1
TEST_CASE_ROOT=$2
TEST_CASE_WORKSPACE="${TEST_CASE_ROOT}/workspace"

LIBRARY_VERSION_UNDER_TEST=$(git log --format="%H" -n 1)
REPOSITORY_UNDER_TEST=${TRAVIS_REPO_SLUG:-SAP/jenkins-library}

git clone -b "${TEST_CASE}" https://github.com/sap/cloud-s4-sdk-book "${TEST_CASE_WORKSPACE}"
cp -f jenkins.yml "${TEST_CASE_WORKSPACE}"
cd "${TEST_CASE_WORKSPACE}" || exit 1

# Configure path to library-repository under test in Jenkins config
sed -i -e "s:__REPO_SLUG__:${REPOSITORY_UNDER_TEST}:g" jenkins.yml

# Force usage of library version under test by setting it in the Jenkinsfile which is then the first definition and thus has the highest precedence
echo "@Library(\"piper-library-os@$LIBRARY_VERSION_UNDER_TEST\") _" | cat - Jenkinsfile > temp && mv temp Jenkinsfile

# Commit the changed version because artifactSetVersion expects the git repo not to be dirty
git commit --all --author="piper-testing-bot <piper-testing-bot@example.com>" --message="Set piper lib version for test"

docker run -v /var/run/docker.sock:/var/run/docker.sock -v "${PWD}":/workspace -v /tmp -e CASC_JENKINS_CONFIG=/workspace/jenkins.yml \
    -e CX_INFRA_IT_CF_USERNAME -e CX_INFRA_IT_CF_PASSWORD -e BRANCH_NAME="${TEST_CASE}" ppiper/jenkinsfile-runner

RC=$?

cd - || { echo "[ERROR] change directory back into integration test root folder failed."; exit 1; }

[ "${RC}" == 0 ] && touch "${TEST_CASE_ROOT}/SUCCESS"
