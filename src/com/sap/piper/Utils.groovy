package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.analytics.Telemetry
import groovy.text.GStringTemplateEngine

import java.nio.charset.StandardCharsets
import java.security.MessageDigest

def stash(Map params) {
    if(params.includes == null) params.includes = '**/*.*'
    if(params.excludes == null) params.excludes = ''
    if(params.useDefaultExcludes == null) params.useDefaultExcludes = true
    if(params.allowEmpty == null) params.allowEmpty = false
    return stash(params.name, params.includes, params.excludes, params.useDefaultExcludes, params.allowEmpty)
}

def stash(String name, String includes = '**/*.*', String excludes = '', boolean useDefaultExcludes = true, boolean allowEmpty = false) {
    if(!name) throw new IllegalArgumentException("name must not be '$name'")
    echo "Stash content: ${name} (includes: ${includes}, excludes: ${excludes}, useDefaultExcludes: ${useDefaultExcludes}, allowEmpty: ${allowEmpty})"

    Map stashParams = [
        name    : name,
        includes: includes,
        excludes: excludes
    ]
    //only set the optional parameter if default excludes should not be applied
    if (!useDefaultExcludes) {
        stashParams.useDefaultExcludes = useDefaultExcludes
    }
    //only set the optional parameter if allow empty should be applied
    if (allowEmpty) {
        stashParams.allowEmpty = allowEmpty
    }
    steps.stash stashParams
}

def stashList(script, List stashes) {
    for (def stash : stashes) {
        def name = stash.name
        def includes = stash.includes
        def excludes = stash.excludes

        if (stash?.merge == true) {
            String lockingResourceGroup = script.commonPipelineEnvironment.projectName?:env.JOB_NAME
            String lockName = "${lockingResourceGroup}/${stash.name}"
            lock(lockName) {
                unstash stash.name
                echo "Stash content: ${name} (includes: ${includes}, excludes: ${excludes})"
                steps.stash name: name, includes: includes, excludes: excludes, allowEmpty: true
            }
        } else {
            echo "Stash content: ${name} (includes: ${includes}, excludes: ${excludes})"
            steps.stash name: name, includes: includes, excludes: excludes, allowEmpty: true
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

def stashStageFiles(Script script, String stageName) {
    List stashes = script.commonPipelineEnvironment.configuration.stageStashes?.get(stageName)?.stashes ?: []

    stashList(script, stashes)

    //NOTE: We do not delete the directory in case Jenkins runs on Kubernetes.
    // deleteDir() is not required in pods, but would be nice to have the same behaviour and leave a clean fileSystem.
    if (!isInsidePod(script)) {
        script.deleteDir()
    }
}

def unstashStageFiles(Script script, String stageName, List stashContent = []) {
    stashContent += script.commonPipelineEnvironment.configuration.stageStashes?.get(stageName)?.unstash ?: []

    script.deleteDir()
    unstashAll(stashContent)

    return stashContent
}

boolean isInsidePod(Script script) {
    return script.env.POD_NAME
}

def unstash(name, msg = "Unstash failed:") {
    def unstashedContent = []
    try {
        echo "Unstash content: ${name}"
        steps.unstash name
        unstashedContent += name
    } catch (e) {
        echo "$msg $name (${e.getMessage()})"
        if (e.getMessage() != null && e.getMessage().contains("JNLP4-connect")) {
            sleep(3) // Wait 3 seconds in case it has been a network hiccup
            try {
                echo "[Retry JNLP4-connect issue] Unstashing content: ${name}"
                steps.unstash name
                unstashedContent += name
            } catch (errRetry) {
                msg = "[Retry JNLP4-connect issue] Unstashing failed:"
                echo "$msg $name (${errRetry.getMessage()})"
            }
        }
    }
    return unstashedContent
}

def unstashAll(stashContent) {
    def unstashedContent = []
    if (stashContent) {
        for (int i = 0; i < stashContent.size(); i++) {
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
        echo "SAP web analytics is disabled. Please remove any remaining use of 'pushToSWA' function!"
    } catch (ignore) {
        // some error occured in telemetry reporting. This should not break anything though.
        echo "[${parameters.step}] Telemetry Report failed: ${ignore.getMessage()}"
    }
}

@NonCPS
static String fillTemplate(String templateText, Map binding) {
    def engine = new GStringTemplateEngine()
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

/*
 * Uses the Maven Help plugin to evaluate the given expression into the resolved values
 * that maven sees at / generates at runtime. This way, the exact Maven coordinates and
 * variables can be used.
 */
static String evaluateFromMavenPom(Script script, String pomFileName, String pomPathExpression) {

    String resolvedExpression = script.mavenExecute(
        script: script,
        pomPath: pomFileName,
        goals: ['org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate'],
        defines: ["-Dexpression=$pomPathExpression", "-DforceStdout", "-q"],
        returnStdout: true
    )
    if (resolvedExpression.startsWith('null object or invalid expression')) {
        // There is no error indication (exit code or otherwise) from the
        // 'evaluate' Maven plugin, only this output to stdout. The calling
        // code assumes an empty string is returned when the property could
        // not be resolved.
        throw new RuntimeException("Cannot evaluate property value from '${pomFileName}', " +
            "missing property or invalid expression '${pomPathExpression}'.")
    }
    return resolvedExpression
}

static List appendParameterToStringList(List list, Map parameters, String paramName) {
    def value = parameters[paramName]
    List result = []
    result.addAll(list)
    if (value in CharSequence) {
        result.add(value)
    } else if (value in List) {
        result.addAll(value)
    }
    return result
}
