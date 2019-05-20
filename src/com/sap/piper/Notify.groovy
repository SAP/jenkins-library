package com.sap.piper

import com.sap.piper.analytics.Telemetry

class Notify implements Serializable {
    protected static enum Severity { ERROR, WARNING }

    protected final static String LIBRARY_NAME = 'piper-lib-os'

    static void warning(Script step, String message, String stepName = null){
        notify(step, message, stepName, Severity.WARNING)
    }

    static void error(Script step, String message, String stepName = null) {
        notify(step, message, stepName, Severity.ERROR)
    }

    private static void notify(Script step, String message, String stepName, Severity severity){
        stepName = stepName ?: step.STEP_NAME

        def notification = "[${severity}] ${message} (${LIBRARY_NAME}/${stepName})"

        if (severity == Severity.ERROR){
            step.error(notification)
        } else{
            step.echo(notification)
        }
    }
}
