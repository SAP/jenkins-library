#!/usr/bin/env bash

# shellcheck disable=SC2002
cat .gitignore | xargs -L1 rm -r
