package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

@NonCPS
def getMandatoryParameter(Map map, paramName, defaultValue = null) {

    def paramValue = map[paramName]

    if (paramValue == null)
        paramValue = defaultValue

    if (paramValue == null)
        throw new Exception("ERROR - NO VALUE AVAILABLE FOR ${paramName}")
    return paramValue

}

def stash(name, include = '**/*.*', exclude = '') {
    echo "Stash content: ${name} (include: ${include}, exclude: ${exclude})"
    steps.stash name: name, includes: include, excludes: exclude
}

def stashWithMessage(name, msg, include = '**/*.*', exclude = '') {
    try {
        stash(name, include, exclude)
    } catch (e) {
        echo msg + name + " (${e.getMessage()})"
    }
}

def unstash(name, msg = "Unstash failed:") {
    def unstashedContent = []
    try {
        echo "Unstash content: ${name}"
        steps.unstash name
        unstashedContent += name
    } catch (e) {
        echo "$msg $name (${e.getMessage()})"
    }
    return unstashedContent
}

def unstashAll(stashContent) {
    def unstashedContent = []
    if (stashContent) {
        for (i = 0; i < stashContent.size(); i++) {
            unstashedContent += unstash(stashContent[i])
        }
    }
    return unstashedContent
}

