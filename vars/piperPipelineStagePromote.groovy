import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.ReportAggregator
import com.sap.piper.StageNameProvider
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String TECHNICAL_STAGE_NAME = 'promote'

@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /** For Docker builds: pushes the Docker image to a container registry. */
    'containerPushToRegistry',
    /** For Maven/MTA builds: uploads artifacts to a Nexus repository manager. */
    'nexusUpload',
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * This stage is responsible to promote build artifacts to an artifact repository / container registry where they can be used from in production deployments.<br />
 *
 */
@GenerateStageDocumentation(defaultStageName = 'Promote')
void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()
    def stageName = StageNameProvider.instance.getStageName(script, parameters, this)

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('containerPushToRegistry', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.containerPushToRegistry)
        .addIfEmpty('nexusUpload', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.nexusUpload)
        .use()

    piperStageWrapper (script: script, stageName: stageName) {
        durationMeasure(script: script, measurementName: 'promote_duration') {
            if (config.containerPushToRegistry) {
                containerPushToRegistry script: script
            }

            if (config.nexusUpload) {
                nexusUpload script: script
                ReportAggregator.instance.reportDeploymentToNexus()
            }
        }
    }
}
