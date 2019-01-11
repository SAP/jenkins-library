#!/bin/bash

d=$(dirname "$0")
[ !  -z  "$d"  ] &&  d="$d/"

export CLASSPATH_FILE='cp.txt'
mvn dependency:build-classpath -Dmdep.outputFile=${CLASSPATH_FILE} > /dev/null 2>&1
groovy -cp "src:$(cat $CLASSPATH_FILE)" "${d}createDocu" "${@}"
