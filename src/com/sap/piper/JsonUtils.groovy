package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

@NonCPS
String groovyObjectToPrettyJsonString(object) {
    return groovy.json.JsonOutput.prettyPrint(groovy.json.JsonOutput.toJson(object))
}

@NonCPS
String groovyObjectToJsonString(object) {
    return groovy.json.JsonOutput.toJson(object)
}

@NonCPS
def jsonStringToGroovyObject(text) {
    return new groovy.json.JsonSlurperClassic().parseText(text)
}
