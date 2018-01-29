#!/bin/bash

echo "Found pull request: Running maven tests and building documentation"

# Run tests
mvn test -B

# Build documentation to check for build errors and warnings
echo "Building documentation to check for errors"
cd $TRAVIS_BUILD_DIR/documentation
mkdocs build --clean --verbose --strict
exit $?
