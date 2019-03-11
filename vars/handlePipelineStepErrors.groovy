import com.cloudbees.groovy.cps.NonCPS

import com.sap.piper.ConfigurationHelper

import groovy.text.SimpleTemplateEngine
import groovy.transform.Field

@Field STEP_NAME = getClass().getName()

@Field Set PARAMETER_KEYS = [
    'echoDetails',
    'stepName',
    'stepNameDoc',
    'stepParameters'
]

void call(Map parameters = [:], body) {
    // load default & individual configuration
    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixin(parameters, PARAMETER_KEYS)
        .withMandatoryProperty('stepParameters')
        .withMandatoryProperty('stepName')
        .addIfEmpty('stepNameDoc' , parameters.stepName)
        .use()

    def message = ''
    try {
        if (config.echoDetails)
            echo "--- Begin library step of: ${config.stepName} ---"

        body()
    } catch (Throwable error) {
        if (config.echoDetails)
            message += formatErrorMessage(config, error)
        throw error
    } finally {
        if (config.echoDetails)
            message += "--- End library step of: ${config.stepName} ---"
        echo message
    }
}

@NonCPS
String formatErrorMessage(Map config, error){
    Map binding = [
        stepName: config.stepName,
        stepParameters: config.stepParameters?.toString(),
        error: error
    ]
    return SimpleTemplateEngine
        .newInstance()
        .createTemplate(libraryResource('com.sap.piper/templates/error.log'))
        .make(binding)
        .toString()
}
