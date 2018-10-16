#!/bin/bash

export CLASSPATH_FILE='cp.txt'
mvn dependency:build-classpath -Dmdep.outputFile=${CLASSPATH_FILE} 2>&1 > /dev/null
#~/Library/groovy-2.4.13/bin/groovy -cp src:`cat $CLASSPATH_FILE` createDocu
groovy -cp src:`cat $CLASSPATH_FILE` createDocu $@
