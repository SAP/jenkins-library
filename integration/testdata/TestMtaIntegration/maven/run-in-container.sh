#!/usr/bin/env bash

set -e
set -x

mkdir mym2

cd /project
/piper mtaBuild --m2Path=mym2

find /project/mym2

[ -f /project/mym2/mygroup/mymvn/1.0-SNAPSHOT/mymvn-1.0-SNAPSHOT.pom ] || (echo "Assertion failed, file mymvn-1.0-SNAPSHOT.pom must exist"; exit 1)
[ -f /project/mym2/mygroup/mymvn-app/1.0-SNAPSHOT/mymvn-app-1.0-SNAPSHOT.war ] || (echo "Assertion failed, file mymvn-app-1.0-SNAPSHOT.war must exist"; exit 1)
[ -f /project/mym2/mygroup/mymvn-app/1.0-SNAPSHOT/mymvn-app-1.0-SNAPSHOT-classes.jar ] || (echo "Assertion failed, file mymvn-app-1.0-SNAPSHOT-classes.jar must exist"; exit 1)
