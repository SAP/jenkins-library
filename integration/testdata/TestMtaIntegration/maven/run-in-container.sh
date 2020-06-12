#!/usr/bin/env bash

set -e
set -x

cd /project
/piper mtaBuild

find /home/mta/.m2/repository/mygroup

[ -f /home/mta/.m2/repository/mygroup/mymvn/1.0-SNAPSHOT/mymvn-1.0-SNAPSHOT.pom ] || (echo "Assertion failed, file mymvn-1.0-SNAPSHOT.pom must exist"; exit 1)
[ -f /home/mta/.m2/repository/mygroup/mymvn-app/1.0-SNAPSHOT/mymvn-app-1.0-SNAPSHOT.war ] || (echo "Assertion failed, file mymvn-app-1.0-SNAPSHOT.war must exist"; exit 1)
[ -f /home/mta/.m2/repository/mygroup/mymvn-app/1.0-SNAPSHOT/mymvn-app-1.0-SNAPSHOT-classes.jar ] || (echo "Assertion failed, file mymvn-app-1.0-SNAPSHOT-classes.jar must exist"; exit 1)
