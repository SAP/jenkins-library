#!/bin/bash

mvn clean test
groovy steps.groovy


[ -d jenkins_home ] && rm -rf jenkins_home
 cp -r jenkins_home_init jenkins_home

CALLS="`pwd`/jenkins_home/piper/calls.json"

mkdir -p `dirname ${CALLS}`

mv target/performedCalls.json ${CALLS}

[ -f ${CALLS} ] || { echo "File \"${CALLS}\" does not exist." ; exit 1; }

cID=$(docker run -d -v `pwd`/jenkins_home:/var/jenkins_home --env calls=/var/jenkins_home/piper/calls.json --env result=/var/jenkins_home/piper/result.json  ppiper/jenkins-master);
echo "ContainerId: ${cID}"; 
while true
do
    [ -f jenkins_home/piper/result.json ] && { docker rm -f ${cID}; break; } # normal ...
    [ -f jenkins_home/piper/FAILURE ] && { docker rm -f ${cID}; break; }     # executing of our init script failed
    docker ps --no-trunc |grep -q ${cID} || break                            # docker container does not run anymore
    echo "[INFO] waiting for results"
    sleep 10
done

RESULT="`pwd`/jenkins_home/piper/result.json"
[ -f ${RESULT} ] && cat jenkins_home/piper/result.json
