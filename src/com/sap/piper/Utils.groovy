package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.analytics.Telemetry
import groovy.text.SimpleTemplateEngine
import hudson.AbortException

import java.nio.charset.StandardCharsets
import java.security.MessageDigest


def stash(name, include = '**/*.*', exclude = '', useDefaultExcludes = true) {
    echo "Stash content: ${name} (include: ${include}, exclude: ${exclude}, useDefaultExcludes: ${useDefaultExcludes})"

    Map stashParams = [
        name    : name,
        includes: include,
        excludes: exclude
    ]
    //only set the optional parameter if default excludes should not be applied
    if (!useDefaultExcludes) {
        stashParams.useDefaultExcludes = useDefaultExcludes
    }
    steps.stash stashParams
}

@NonCPS
def runClosures(Map closures) {

    def closuresToRun = closures.values().asList()
    Collections.shuffle(closuresToRun) // Shuffle the list so no one tries to rely on the order of execution
    for (int i = 0; i < closuresToRun.size(); i++) {
        (closuresToRun[i] as Closure).run()
    }
}

static runWithPostAction(Script context, Closure action, Closure postAction = null) {

    Exception exAction, exPostAction

    def result

    try {
        result = action()
    } catch(Exception e) {
        exAction = e
    } finally {
        try {
            if(postAction)
                postAction()
        } catch(Exception e) {
            exPostAction = e
        }
    }
    if(exAction) {
        if(exPostAction) {
            // [Q] What is the reason for the echo statement below?
            // [A] In case the exception raised by the action is a hudson.AbortException we only see the message from that exception in the log - no stacktrace.
            //     Hence the suppressed exception - which has been in fact added to that hudson.AbortException is not visible.
            if(exAction instanceof AbortException ) {
                context.echo "Got an '${exAction.class.name}' from the action and an '${exPostAction.class.name}' from the post action. " +
                "The exception from the post action was: '${exPostAction}'."
            }
            exAction.addSuppressed(exPostAction)
        }
        throw exAction
    }
    if(exPostAction) {
        throw exPostAction
    }

    result
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
            if (stashContent[i]) {
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
        parameters.jobUrlSha1 = generateSha1(env.JOB_URL)
        parameters.buildUrlSha1 = generateSha1(env.BUILD_URL)

        Telemetry.notify(this, config, parameters)
    } catch (ignore) {
        // some error occured in telemetry reporting. This should not break anything though.
    }
}

@NonCPS
static String fillTemplate(String templateText, Map binding) {
    def engine = new SimpleTemplateEngine()
    String result = engine.createTemplate(templateText).make(binding)
    return result
}

static String downloadSettingsFromUrl(script, String url, String targetFile = 'settings.xml') {
    if (script.fileExists(targetFile)) {
        throw new RuntimeException("Trying to download settings file to ${targetFile}, but a file with this name already exists. Please specify a unique file name.")
    }

    def settings = script.httpRequest(url)
    script.writeFile(file: targetFile, text: settings.getContent())
    return targetFile
}
