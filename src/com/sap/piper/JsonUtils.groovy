package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS
import groovy.json.JsonSlurperClassic

@NonCPS
String getPrettyJsonString(object) {
    return groovy.json.JsonOutput.prettyPrint(groovy.json.JsonOutput.toJson(object))
}

@NonCPS
def parseJsonSerializable(text) {
    return new JsonSlurperClassic().parseText(text)
}
