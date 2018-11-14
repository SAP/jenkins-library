package com.sap.piper

import static java.lang.Boolean.getBoolean

import static com.sap.piper.JenkinsUtils.isPluginActive

import com.sap.piper.Dependencies

static checkScript(def step, Map params) {

    def script = params?.script

    if(script == null) {

        if(getBoolean('com.sap.piper.featureFlag.failOnMissingScript')) {
            step.error("[ERROR][${step.STEP_NAME}] No reference to surrounding script provided with key 'script', e.g. 'script: this'.")
        }

        step.currentBuild.setResult('UNSTABLE')

        step.echo "[WARNING][${step.STEP_NAME}] No reference to surrounding script provided with key 'script', e.g. 'script: this'. " +
                    "Build status has been set to 'UNSTABLE'. In future versions of piper-lib the build will fail."
    }

    return script
}

static void checkRequiredPlugins(step) {

    def requiredPlugins = step.getClass().getDeclaredMethod('call')?.getAnnotation(Dependencies)?.requiredPlugins() ?: []

    def missingPlugins = []

    for (plugin in requiredPlugins)
        if(! isPluginActive(plugin)) missingPlugins << plugin

    if(missingPlugins) step.error "The following plugins are required for step '${step.STEP_NAME}', but they are not available: ${missingPlugins}."
}
