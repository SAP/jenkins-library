#!/usr/bin/env bash

EXAMPLE_PROJECT_BRANCH=$1

LIBRARY_VERSION_UNDER_TEST=$(git log --format="%H" -n 1)
REPOSITORY_UNDER_TEST=${TRAVIS_REPO_SLUG:-SAP/jenkins-library}

rm -rf workspace
git clone -b "${EXAMPLE_PROJECT_BRANCH}" https://github.com/sap/cloud-s4-sdk-book workspace
cp -f ../jenkins.yml workspace
cd workspace || exit 1

# Configure path to library-repository under test in Jenkins config
sed -i -e "s:__REPO_SLUG__:${REPOSITORY_UNDER_TEST}:g" jenkins.yml

# Force usage of library version under test by setting it in the Jenkinsfile which is then the first definition and thus has the highest precedence
echo "@Library(\"piper-library-os@$LIBRARY_VERSION_UNDER_TEST\") _" | cat - Jenkinsfile > temp && mv temp Jenkinsfile

# Commit the changed version because artifactSetVersion expects the git repo not to be dirty
git commit --all --author="piper-testing-bot <piper-testing-bot@example.com>" --message="Set piper lib version for test"

docker run -v /var/run/docker.sock:/var/run/docker.sock -v "${PWD}":/workspace -v /tmp -e CASC_JENKINS_CONFIG=/workspace/jenkins.yml \
    -e CX_INFRA_IT_CF_USERNAME -e CX_INFRA_IT_CF_PASSWORD -e BRANCH_NAME="${EXAMPLE_PROJECT_BRANCH}" ppiper/jenkinsfile-runner
