import com.cloudbees.groovy.cps.NonCPS

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.analytics.InfluxData

import groovy.text.SimpleTemplateEngine
import groovy.transform.Field
import hudson.AbortException

import org.jenkinsci.plugins.workflow.steps.FlowInterruptedException

@Field STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Defines the behavior, in case an error occurs which is handled by this step. When set to `false` an error results in an "UNSTABLE" build result and the pipeline can continue.
     * @possibleValues `true`, `false`
     */
    'failOnError',
    /** Defines the url of the library's documentation that will be used to generate the corresponding links to the step documentation.*/
    'libraryDocumentationUrl',
    /** Defines the url of the library's repository that will be used to generate the corresponding links to the step implementation.*/
    'libraryRepositoryUrl',
    /** Defines a list of mandatory steps (step names) which have to be successful (=stop the pipeline), even if `failOnError: false` */
    'mandatorySteps',
    /**
     * Defines a Map containing step name as key and timout in minutes in order to stop an execution after a certain timeout.
     * This helps to make pipeline runs more resilient with respect to long running steps.
     * */
    'stepTimeouts'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /**
     * If it is set to true details will be output to the console. See example below.
     * @possibleValues `true`, `false`
     */
    'echoDetails',
    /** Defines the name of the step for which the error handling is active. It will be shown in the console log.*/
    'stepName',
    /** Defines the documented step, in case the documentation reference should point to a different step. */
    'stepNameDoc',
    /** Passes the parameters of the step which uses the error handling onto the error handling. The list of parameters is then shown in the console output.*/
    'stepParameters'
])

/**
 * Used by other steps to make error analysis easier. Lists parameters and other data available to the step in which the error occurs.
 */
@GenerateDocumentation
void call(Map parameters = [:], body) {
    // load default & individual configuration
    def cpe = parameters.stepParameters?.script?.commonPipelineEnvironment ?: null
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
        if (!config.failOnError && config.stepTimeouts?.get(config.stepName)) {
            timeout(time: config.stepTimeouts[config.stepName]) {
                body()
            }
        } else {
            body()
        }
    } catch (AbortException | FlowInterruptedException ex) {
        if (config.echoDetails)
            message += formatErrorMessage(config, ex)
        writeErrorToInfluxData(config, ex)
        if (config.failOnError || config.stepName in config.mandatorySteps) {
            throw ex
        }

        if (config.stepParameters?.script) {
            config.stepParameters?.script.currentBuild.result = 'UNSTABLE'
        } else {
            currentBuild.result = 'UNSTABLE'
        }

        echo "[${STEP_NAME}] Error in step ${config.stepName} - Build result set to 'UNSTABLE'"

        List unstableSteps = cpe?.getValue('unstableSteps') ?: []
        if(!unstableSteps) {
            unstableSteps = []
        }

        // add information about unstable steps to pipeline environment
        // this helps to bring this information to users in a consolidated manner inside a pipeline
        unstableSteps.add(config.stepName)
        cpe?.setValue('unstableSteps', unstableSteps)

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
    if(InfluxData.getInstance().getFields().pipeline_data?.build_error_message == null){
        InfluxData.addTag('pipeline_data', 'build_error_step', config.stepName)
        InfluxData.addTag('pipeline_data', 'build_error_stage', config.stepParameters.script?.env?.STAGE_NAME)
        InfluxData.addField('pipeline_data', 'build_error_message', error.getMessage())
    }
}
