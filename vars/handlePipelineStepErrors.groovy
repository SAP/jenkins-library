import com.cloudbees.groovy.cps.NonCPS

import com.sap.piper.ConfigurationHelper
import com.sap.piper.analytics.InfluxData

import groovy.text.SimpleTemplateEngine
import groovy.transform.Field

@Field STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = []
@Field Set PARAMETER_KEYS = [
    'echoDetails',
    'libraryDocumentationUrl',
    'libraryRepositoryUrl',
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
        writeErrorToInfluxData(config, error)
        throw error
    } finally {
        if (config.echoDetails)
            message += "--- End library step of: ${config.stepName} ---"
        echo message
    }
}

@NonCPS
private String formatErrorMessage(Map config, error){
    Map binding = [
        error: error,
        libraryDocumentationUrl: config.libraryDocumentationUrl,
        libraryRepositoryUrl: config.libraryRepositoryUrl,
        stepName: config.stepName,
        stepParameters: config.stepParameters?.toString()
    ]
    return SimpleTemplateEngine
        .newInstance()
        .createTemplate(libraryResource('com.sap.piper/templates/error.log'))
        .make(binding)
        .toString()
}

private void writeErrorToInfluxData(Map config, error){
    def script = config?.stepParameters?.script

    if(InfluxData.getInstance().getTags().build_error_message == null){
        InfluxData.addTag('pipeline_data', 'build_error_step', config.stepName)
        InfluxData.addTag('pipeline_data', 'build_error_stage', script.env?.STAGE_NAME)
        InfluxData.addField('pipeline_data', 'build_error_message', error.getMessage())
    }
}
