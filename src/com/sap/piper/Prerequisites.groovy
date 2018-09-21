package com.sap.piper

import static java.lang.Boolean.getBoolean

/**
  * @return <code>true</code> in case the script has been provided. <code>false</code> otherwise.
  */
static checkScript(def step, Map params) {

    def script = params?.script

    if(script == null) {

        step.currentBuild.status = 'UNSTABLE'

        step.echo "[WARNING][${step.STEP_NAME}] No reference to surrounding script provided with key 'script', e.g. 'script: this'. " +
                   "Build status has been set to '${step.currentBuild.status}'. In future versions of piper-lib the build will fail."

        if(getBoolean('com.sap.piper.featureFlag.failOnMissingScript')) {
            step.error("[ERROR][${step.STEP_NAME}] No reference to surrounding script provided with key 'script', e.g. 'script: this'.")
        }


    }

    return script
}
