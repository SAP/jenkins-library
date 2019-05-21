#!/bin/bash

d=$(dirname "$0")
[ !  -z  "$d"  ] &&  d="$d/"

export CLASSPATH_FILE='target/cp.txt'
mvn clean test dependency:build-classpath -Dmdep.outputFile=${CLASSPATH_FILE} > /dev/null 2>&1
groovy  "${d}steps" -in target/trackedCalls.json --out target/performedCalls.json
groovy -cp "target/classes:$(cat $CLASSPATH_FILE)" "${d}createDocu" "${@}"
