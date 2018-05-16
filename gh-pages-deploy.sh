#!/bin/bash

echo "Found change on master: Deployment of documentation"

PRIVATE_KEY="cfg/id_rsa"

chmod 600 "${PRIVATE_KEY}"
eval `ssh-agent -s`
ssh-add "${PRIVATE_KEY}"
git config user.name "Travis CI Publisher"
git remote add docu "git@github.com:$TRAVIS_REPO_SLUG.git";
git fetch docu gh-pages:gh-pages
echo "Pushing to gh-pages of repository $TRAVIS_REPO_SLUG"
cd $TRAVIS_BUILD_DIR/documentation
mkdocs gh-deploy -v --clean --remote-name docu
