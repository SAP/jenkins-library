package com.sap.piper

class Notify implements Serializable {
    private static enum Severity { ERROR, WARNING }
    private final static String LIBRARY_NAME = 'piper-lib-os'

    protected static Utils utils = null

    protected static Utils getUtilsInstance(){
        this.utils = this.utils ?: new Utils()
        return this.utils
    }

    static void warning(Map config, Script step, String message, String stepName = null){
        log(config, step, message, stepName, Severity.WARNING)
    }

    static void error(Map config, Script step, String message, String stepName = null) {
        log(config, step, message, stepName, Severity.ERROR)
    }

    private static void log(Map config, Script step, String message, String stepName, Severity severity){
        stepName = stepName ?: step.STEP_NAME
        getUtilsInstance().pushToSWA([
            folder: '',
            repository: '',
            step: 'Notify',
            eventType: 'notification',
            stepParam1: LIBRARY_NAME,
            stepParam2: stepName,
            stepParam3: message,
            stepParam4: severity
        ], config)

        def notification = "[${severity}] ${message} (${LIBRARY_NAME}/${stepName})"

        if (severity == Severity.ERROR){
            step.error(notification)
        }
        step.echo(notification)
    }
}
