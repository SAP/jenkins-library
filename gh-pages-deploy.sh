#!/bin/bash

echo "Found change on master: Deployment of documentation"

. .travis.env

# We can't deploy if any of these vars are empty
for variable in "$KEY" "$IV" "$ENCRYPTED_FILE"; do
    if [ -z "$variable" ]; then
        echo "ERROR: Found empty KEY, IV, or ENCRYPTED_FILE variable. Exiting."
        exit 1
    fi
done

openssl aes-256-cbc -K "$KEY" -iv "$IV" -in "$ENCRYPTED_FILE" -out deploy-key -d

chmod 600 deploy-key
eval `ssh-agent -s`
ssh-add deploy-key
git config user.name "Travis CI Publisher"
git remote add gh-token "git@github.com:$TRAVIS_REPO_SLUG.git";
git fetch gh-token && git fetch gh-token gh-pages:gh-pages
echo "Pushing to gh-pages of repository $TRAVIS_REPO_SLUG"
cd $TRAVIS_BUILD_DIR/documentation
mkdocs gh-deploy -v --clean --remote-name gh-token
