package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.analytics.Telemetry
import groovy.text.SimpleTemplateEngine

import java.nio.charset.StandardCharsets
import java.security.MessageDigest

@NonCPS
def getMandatoryParameter(Map map, paramName, defaultValue = null) {

    def paramValue = map[paramName]

    if (paramValue == null)
        paramValue = defaultValue

    if (paramValue == null)
        throw new Exception("ERROR - NO VALUE AVAILABLE FOR ${paramName}")
    return paramValue

}

def stash(name, include = '**/*.*', exclude = '', useDefaultExcludes = true) {
    echo "Stash content: ${name} (include: ${include}, exclude: ${exclude}, useDefaultExcludes: ${useDefaultExcludes})"

    Map stashParams = [
        name: name,
        includes: include,
        excludes: exclude
    ]
    //only set the optional parameter if default excludes should not be applied
    if (!useDefaultExcludes) {
        stashParams.useDefaultExcludes = useDefaultExcludes
    }
    steps.stash stashParams
}

def stashList(script, List stashes) {
    for (def stash : stashes) {
        def name = stash.name
        def include = stash.includes
        def exclude = stash.excludes

        if (stash?.merge == true) {
            String lockName = "${script.commonPipelineEnvironment.configuration.stashFiles}/${stash.name}"
            lock(lockName) {
                unstash stash.name
                echo "Stash content: ${name} (include: ${include}, exclude: ${exclude})"
                steps.stash name: name, includes: include, exclude: exclude, allowEmpty: true
            }
        } else {
            echo "Stash content: ${name} (include: ${include}, exclude: ${exclude})"
            steps.stash name: name, includes: include, exclude: exclude, allowEmpty: true
        }
    }
}

def stashWithMessage(name, msg, include = '**/*.*', exclude = '', useDefaultExcludes = true) {
    try {
        stash(name, include, exclude, useDefaultExcludes)
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
            if(stashContent[i]) {
                unstashedContent += unstash(stashContent[i])
            }
        }
    }
    return unstashedContent
}

@NonCPS
def generateSha1(input) {
    return MessageDigest
        .getInstance("SHA-1")
        .digest(input.getBytes(StandardCharsets.UTF_8))
        .encodeHex().toString()
}

void pushToSWA(Map parameters, Map config) {
    try {
        parameters.actionName = parameters.get('actionName') ?: 'Piper Library OS'
        parameters.eventType = parameters.get('eventType') ?: 'library-os'
        parameters.jobUrlSha1 =  generateSha1(env.JOB_URL)
        parameters.buildUrlSha1 = generateSha1(env.BUILD_URL)

        Telemetry.notify(this, config, parameters)
    } catch (ignore) {
        // some error occured in telemetry reporting. This should not break anything though.
    }
}

@NonCPS
static String fillTemplate(String templateText, Map binding){
    def engine = new SimpleTemplateEngine()
    String result = engine.createTemplate(templateText).make(binding)
    return result
}
