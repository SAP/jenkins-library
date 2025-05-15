import static com.sap.piper.Prerequisites.checkScript
import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = [
    /**
     * Name of the docker image that should be used, in which node should be installed and configured. Default value is 'dlang2/dmd-ubuntu:latest'.
     */
    'dockerImage',
    /** @see dockerExecute*/
    'dockerEnvVars',
    /** @see dockerExecute */
    'dockerOptions',
    /** @see dockerExecute*/
    'dockerWorkspace',
    /**
     * URL of default DUB registry
     */
    'defaultDubRegistry',
    /**
     * Which DUB command should be executed.
     */
    'dubCommand']
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS + [
    /**
     * Docker options to be set when starting the container.
     */
    'dockerOptions']
/**
 * Executes DUB commands inside a docker container.
 * Docker image, docker options and dub commands can be specified or configured.
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

        if (!fileExists('dub.json') && !fileExists('dub.sdl')) {
            error "[${STEP_NAME}] Neither dub.json nor dub.sdl was found."
        }
        dockerExecute(script: script,
            dockerImage: configuration.dockerImage,
            dockerEnvVars: configuration.dockerEnvVars,
            dockerOptions: configuration.dockerOptions,
            dockerWorkspace: configuration.dockerWorkspace
        ) {
            if (configuration.defaultDubRegistry) {
                sh """
                    mkdir ~/.dub
                    echo '{"skipRegistry": "standard", "registryUrls": ["${configuration.defaultDubRegistry}"]}' > ~/.dub/settings.json
                """
            }
            if (configuration.dubCommand) {
                sh "dub ${configuration.dubCommand}"
            }
            if (body) {
                body()
            }
        }
    }
}
