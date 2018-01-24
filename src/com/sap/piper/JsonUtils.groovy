package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

@NonCPS
def getPrettyJsonString(object) {
    return groovy.json.JsonOutput.prettyPrint(groovy.json.JsonOutput.toJson(object))
}
