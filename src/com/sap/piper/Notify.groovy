package com.sap.piper

import groovy.text.SimpleTemplateEngine

class Notify implements Serializable {
    private static enum Severity { ERROR, WARNING }
    private final static String LIBRARY_NAME = 'piper-lib-os'
    private final static String MESSAGE_PATTERN = '[${severity}] ${message} (${libName}/${stepName})'

    protected static Utils instance = null

    protected static Utils getUtilsInstance(){
        instance = instance ?: new Utils()
        return instance
    }

    static void warning(Map config, Script step, String msg, String stepName = null){
        log(config, step, msg, stepName)
    }

    static void error(Map config, Script step, String msg, String stepName = null) {
        log(config, step, msg, stepName, Severity.ERROR)
    }

    private static void log(Map config, Script step, String msg, String stepName, Severity severity = Severity.WARNING){
        stepName = stepName ?: step.STEP_NAME
        getUtilsInstance().pushToSWA([
            folder: '',
            repository: '',
            step: 'Notify',
            eventType: 'notification',
            stepParam1: LIBRARY_NAME,
            stepParam2: stepName,
            stepParam3: msg,
            stepParam4: severity
        ], config)
        def logEntry = SimpleTemplateEngine.newInstance().createTemplate(
            MESSAGE_PATTERN
        ).make([
            libName: LIBRARY_NAME,
            stepName: stepName,
            message: msg,
            severity: severity
        ]).toString()

        if (severity == Severity.ERROR){
            step.error(logEntry)
        }
        step.echo(logEntry)
    }
}
