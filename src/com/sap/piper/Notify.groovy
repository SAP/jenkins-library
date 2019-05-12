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

    static void warning(Boolean collectTelemetryData, Script step, String message, String stepName = null){
        notify(collectTelemetryData, step, message, stepName, Severity.WARNING)
    }

    static void error(Boolean collectTelemetryData, Script step, String message, String stepName = null) {
        notify(collectTelemetryData, step, message, stepName, Severity.ERROR)
    }

    private static void notify(Boolean collectTelemetryData, Script step, String message, String stepName, Severity severity){
        stepName = stepName ?: step.STEP_NAME

        Telemetry.notify(step, collectTelemetryData, [
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
