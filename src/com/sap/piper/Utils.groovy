package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS
import org.jenkinsci.plugins.workflow.steps.MissingContextVariableException

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

def generateSha1Inline(input) {
    return "`echo -n '${input}' | sha1sum | sed 's/  -//'`"
}

void pushToSWA(Map parameters, Map config) {
    try {
        //allow opt-out via configuration
        if (!config.collectTelemetryData) {
            return
        }

        def swaCustom = [:]

        /* SWA custom parameters:
            custom3 = step name (passed as parameter step)
            custom4 = job url hashed (calculated)
            custom5 = build url hashed (calculated)
            custom11 = step related parameter 1 (passed as parameter stepParam1)
            custom12 = step related parameter 2 (passed as parameter stepParam2)
            custom13 = step related parameter 3 (passed as parameter stepParam3)
            custom14 = step related parameter 4 (passed as parameter stepParam4)
            custom15 = step related parameter 5 (passed as parameter stepParam5)
        */

        def swaUrl = 'https://webanalytics.cfapps.eu10.hana.ondemand.com/tracker/log'
        def action_name = 'Piper Library OS'
        def idsite = '827e8025-1e21-ae84-c3a3-3f62b70b0130'
        def url = 'https://github.com/SAP/jenkins-library'
        def event_type = 'library'

        swaCustom.custom3 = parameters.get('step')
        swaCustom.custom4 = generateSha1Inline(env.JOB_URL)
        swaCustom.custom5 = generateSha1Inline(env.BUILD_URL)
        swaCustom.custom11 = parameters.get('stepParam1')
        swaCustom.custom12 = parameters.get('stepParam2')
        swaCustom.custom13 = parameters.get('stepParam3')
        swaCustom.custom14 = parameters.get('stepParam4')
        swaCustom.custom15 = parameters.get('stepParam5')

        def options = []
        options.push("-G")
        options.push("-v \"${swaUrl}\"")
        options.push("--data-urlencode \"action_name=${action_name}\"")
        options.push("--data-urlencode \"idsite=${idsite}\"")
        options.push("--data-urlencode \"url=${url}\"")
        options.push("--data-urlencode \"event_type=${event_type}\"")
        for(def key : ['custom3', 'custom4', 'custom5', 'custom11', 'custom12', 'custom13', 'custom14', 'custom15']){
            if (swaCustom[key] != null) options.push("--data-urlencode \"${key}=${swaCustom[key]}\"")
        }
        options.push("--connect-timeout 5")
        options.push("--max-time 20")

        sh(returnStatus: true, script: "#!/bin/sh +x\ncurl ${options.join(' ')} > /dev/null 2>&1 || echo '[${parameters.get('step')}] Telemetry Report to SWA failed!'")
    } catch (MissingContextVariableException noNode) {
        echo "[${parameters.get('step')}] Telemetry Report to SWA skipped, no node available!"
    } catch (ignore) {
        // some error occured in SWA reporting. This should not break anything though.
    }
}

