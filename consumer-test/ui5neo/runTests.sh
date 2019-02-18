#!/bin/bash -e

LIBRARY_VERSION_UNDER_TEST=$(git log --format="%H" -n 1)
REPOSITORY_UNDER_TEST=${TRAVIS_REPO_SLUG:-SAP/jenkins-library}

rm -rf workspace
git clone -b master https://github.com/sbmaier/openui5-sample-app.git workspace
cp -f jenkins.yml workspace
cd workspace

#sed -i -e "s:__REPO_SLUG__:${REPOSITORY_UNDER_TEST}:g" jenkins.yml
#echo "@Library(\"piper-lib-os@$LIBRARY_VERSION_UNDER_TEST\") _" | cat - Jenkinsfile > temp && mv temp Jenkinsfile
git add jenkins.yml
git commit --all --author="piper-testing-bot <null@null.com>" --message="Set piper lib version for test"

docker run -v /var/run/docker.sock:/var/run/docker.sock -v "${PWD}":/workspace -v /tmp -e CASC_JENKINS_CONFIG=/workspace/jenkins.yml -e CX_INFRA_IT_CF_USERNAME -e CX_INFRA_IT_CF_PASSWORD -e BRANCH_NAME=consumer-test -e LIBRARY_VERSION_UNDER_TEST=${LIBRARY_VERSION_UNDER_TEST} -e REPOSITORY_UNDER_TEST=${REPOSITORY_UNDER_TEST} ppiper/jenkinsfile-runner
#docker run -v /var/run/docker.sock:/var/run/docker.sock -v "${PWD}":/workspace -v /tmp -e CASC_JENKINS_CONFIG=/workspace/jenkins.yml -e CX_INFRA_IT_CF_USERNAME -e CX_INFRA_IT_CF_PASSWORD -e BRANCH_NAME=consumer-test  ppiper/jenkinsfile-runner

