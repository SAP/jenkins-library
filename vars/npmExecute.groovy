import static com.sap.piper.Prerequisites.checkScript
import static com.sap.piper.BashUtils.quoteAndEscape as q

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = [
    /**
     * Name of the docker image that should be used, in which node should be installed and configured.
     */
    'dockerImage',
    /** @see dockerExecute*/
    'dockerEnvVars',
    /** @see dockerExecute */
    'dockerOptions',
    /** @see dockerExecute*/
    'dockerWorkspace',
    /**
     * URL of default NPM registry
     */
    'defaultNpmRegistry',
    /**
     * Which NPM command should be executed.
     */
    'npmCommand']
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS + [
    /**
     * Docker options to be set when starting the container.
     */
    'dockerOptions']
/**
 * Executes NPM commands inside a docker container.
 * Docker image, docker options and npm commands can be specified or configured.
 */
@GenerateDocumentation
void call(Map parameters = [:], body = null) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        final script = checkScript(this, parameters) ?: this
        String stageName = parameters.stageName ?: env.STAGE_NAME

        // load default & individual configuration
        Map configuration = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        try {
            if (!fileExists('package.json')) {
                error "[${STEP_NAME}] package.json is not found."
            }
            dockerExecute(script: script,
                dockerImage: configuration.dockerImage,
                dockerEnvVars: configuration.dockerEnvVars,
                dockerOptions: configuration.dockerOptions,
                dockerWorkspace: configuration.dockerWorkspace
            ) {
                sh returnStatus: true, script: """
                    node --version
                    npm --version
                """
                if (configuration.defaultNpmRegistry) {
                    sh "npm config set registry ${q(configuration.defaultNpmRegistry)}"
                }
                if (configuration.npmCommand) {
                    sh "npm ${configuration.npmCommand}"
                }
                if (body) {
                    body()
                }
            }
        } catch (Exception e) {
            println "Error while executing npm. Here are the logs:"
            sh "cat ~/.npm/_logs/*"
            throw e
        }
    }
}
