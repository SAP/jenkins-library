import com.cloudbees.groovy.cps.NonCPS

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper

import groovy.text.SimpleTemplateEngine
import groovy.transform.Field

@Field STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = []
@Field Set PARAMETER_KEYS = [
    /**
     * If set to true the following will be output to the console:
     * 1. Step beginning: `--- Begin library step: ${stepName}.groovy ---`
     * 2. Step end: `--- End library step: ${stepName}.groovy ---`
     * 3. Step errors:
     *
     * ```log
     * ----------------------------------------------------------
     * --- An error occurred in the library step: ${stepName}
     * ----------------------------------------------------------
     * The following parameters were available to the step:
     * ***
     * ${stepParameters}
     * ***
     * The error was:
     * ***
     * ${err}
     * ***
     * Further information:
     * * Documentation of step ${stepName}: .../${stepName}/
     * * Pipeline documentation: https://...
     * * GitHub repository for pipeline steps: https://...
     * ----------------------------------------------------------
     * ```
     * @possibleValues `true`, `false`
     */
    'echoDetails',
    /** Defines the url of the library's documentation that will be used to generate the corresponding links to the step documentation.*/
    'libraryDocumentationUrl',
    /** Defines the url of the library's repository that will be used to generate the corresponding links to the step implementation.*/
    'libraryRepositoryUrl',
    /** Defines the name of the step executed that will be shown in the console output.*/
    'stepName',
    /** */
    'stepNameDoc',
    /** Defines the parameters from the step to be executed. The list of parameters is then shown in the console output.*/
    'stepParameters'
]

/**
 * Used by other steps to make error analysis easier. Lists parameters and other data available to the step in which the error occurs.
 */
@GenerateDocumentation
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

    if(script && script.commonPipelineEnvironment?.getInfluxCustomDataMapTags().build_error_message == null){
        script.commonPipelineEnvironment?.setInfluxCustomDataMapTagsEntry('pipeline_data', 'build_error_step', config.stepName)
        script.commonPipelineEnvironment?.setInfluxCustomDataMapTagsEntry('pipeline_data', 'build_error_stage', script.env?.STAGE_NAME)
        script.commonPipelineEnvironment?.setInfluxCustomDataMapEntry('pipeline_data', 'build_error_message', error.getMessage())
    }
}
