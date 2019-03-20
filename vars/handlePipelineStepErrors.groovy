import com.cloudbees.groovy.cps.NonCPS

import com.sap.piper.ConfigurationHelper

import groovy.text.SimpleTemplateEngine
import groovy.transform.Field
import hudson.AbortException

@Field STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'failOnError',
    'mandatorySteps'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    'echoDetails',
    'libraryDocumentationUrl',
    'libraryRepositoryUrl',
    'stepName',
    'stepNameDoc',
    'stepParameters'
])

void call(Map parameters = [:], body) {
    // load default & individual configuration
    def cpe = parameters.stepParameters?.script?.commonPipelineEnvironment ?: commonPipelineEnvironment
    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(cpe, GENERAL_CONFIG_KEYS)
        .mixinStepConfig(cpe, STEP_CONFIG_KEYS)
        .mixinStageConfig(cpe, parameters.stepParameters?.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
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
    } catch (AbortException ae) {
        if (config.echoDetails)
            message += formatErrorMessage(config, error)
        writeErrorToInfluxData(config, error)
        if (config.failOnError || config.stepName in config.mandatorySteps) {
            throw ae
        }
        if (config.stepParameters?.script) {
            config.stepParameters?.script.currentBuild.result = 'UNSTABLE'
        } else {
            currentBuild.result = 'UNSTABLE'
        }

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

    if(script && script.commonPipelineEnvironment?.getInfluxCustomDataMapTags().build_error_message == null){
        script.commonPipelineEnvironment?.setInfluxCustomDataMapTagsEntry('pipeline_data', 'build_error_step', config.stepName)
        script.commonPipelineEnvironment?.setInfluxCustomDataMapTagsEntry('pipeline_data', 'build_error_stage', script.env?.STAGE_NAME)
        script.commonPipelineEnvironment?.setInfluxCustomDataMapEntry('pipeline_data', 'build_error_message', error.getMessage())
    }
}
