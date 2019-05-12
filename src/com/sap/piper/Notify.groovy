package com.sap.piper

import com.sap.piper.analytics.Telemetry

class Notify implements Serializable {
    protected static enum Severity { ERROR, WARNING }

    protected final static String LIBRARY_NAME = 'piper-lib-os'
    protected static Utils utils = null

    protected static Utils getUtilsInstance(){
        this.utils = this.utils ?: new Utils()
        return this.utils
    }

    static void warning(Map config, Script step, String message, String stepName = null){
        notify(config, step, message, stepName, Severity.WARNING)
    }

    static void error(Map config, Script step, String message, String stepName = null) {
        notify(config, step, message, stepName, Severity.ERROR)
    }

    private static void notify(Map config, Script step, String message, String stepName, Severity severity){
        stepName = stepName ?: step.STEP_NAME

        Telemetry.notify(step, config, [
            folder: '',
            repository: '',
            step: 'Notify',
            actionName: 'Piper Library OS',
            eventType: 'notification',
            jobUrlSha1: Utils.generateSha1(env.JOB_URL),
            buildUrlSha1: Utils.generateSha1(env.BUILD_URL),
            stepParam1: LIBRARY_NAME,
            stepParam2: stepName,
            stepParam3: message,
            stepParam4: severity
        ])

        def notification = "[${severity}] ${message} (${LIBRARY_NAME}/${stepName})"

        if (severity == Severity.ERROR){
            step.error(notification)
        } else{
            step.echo(notification)
        }
    }
}
