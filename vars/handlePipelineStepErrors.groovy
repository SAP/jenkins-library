import com.cloudbees.groovy.cps.NonCPS

import com.sap.piper.ConfigurationHelper

import groovy.text.SimpleTemplateEngine
import groovy.transform.Field
import hudson.AbortException

@Field STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Defines the behavior, in case an error occurs which is handled by this step. When set to `false` an error results in an "UNSTABLE" build result and the pipeline can continue.
     * @possibleValues `true`, `false`
     */
    'failOnError',
    /** Defines a list of mandatory steps (step names) which have to be successful (=stop the pipeline), even if `failOnError: false` */
    'mandatorySteps'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /**
     * Specifies if error details should be printed into the console log.
     * @possibleValues `true`, `false`
     */
    'echoDetails',
    /** This parameter can be used to change the root path of the Library documentation. */
    'libraryDocumentationUrl',
    /** This parameter can be used to change the root path of the Library repository. */
    'libraryRepositoryUrl',
    /** Defines the name of the step for which the error handling is active. It will be shown in the console log.*/
    'stepName',
    /** Defines the documented step, in case the documentation reference should point to a different step. */
    'stepNameDoc',
    /** Passes the parameters of the step which uses the error handling onto the error handling. The list of parameters is then shown in the console output.*/
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
            message += formatErrorMessage(config, ae)
        writeErrorToInfluxData(config, ae)
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
