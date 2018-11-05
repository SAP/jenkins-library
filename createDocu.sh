#!/bin/bash

export CLASSPATH_FILE='cp.txt'
mvn dependency:build-classpath -Dmdep.outputFile=${CLASSPATH_FILE} > /dev/null 2>&1
#~/Library/groovy-2.4.13/bin/groovy -cp src:`cat $CLASSPATH_FILE` createDocu
groovy -cp "src:$(cat $CLASSPATH_FILE)" createDocu "${@}"
