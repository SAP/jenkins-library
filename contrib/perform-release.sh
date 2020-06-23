#!/bin/sh -e

# Manually trigger a release of project "Piper".
# Usually we do release on a schedule, but sometimes you might need to trigger a release.
# Invoke this script with PIPER_RELEASE_TOKEN set to your personal access token for GitHub with 'repo' scope.
# This script is based on https://goobar.io/2019/12/07/manually-trigger-a-github-actions-workflow/

if [ -z "$PIPER_RELEASE_TOKEN" ]
then
    echo "Required variable PIPER_RELEASE_TOKEN is not set, please set a personal access token for GitHub with 'repo' scope."
    exit 1
fi

curl -H "Accept: application/vnd.github.everest-preview+json" \
    -H "Authorization: token ${PIPER_RELEASE_TOKEN}" \
    --request POST \
    --data '{"event_type": "perform-release"}' \
    https://api.github.com/repos/SAP/jenkins-library/dispatches
