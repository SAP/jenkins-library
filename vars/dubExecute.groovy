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

        // load default & individual configuration
        Map configuration = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], configuration)

        if (!fileExists('dub.json') && !fileExists('dub.sdl')) {
            error "[${STEP_NAME}] Neither dub.json nor dub.sdl was found."
        }
        dockerExecute(script: script, dockerImage: configuration.dockerImage, dockerOptions: configuration.dockerOptions) {
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
