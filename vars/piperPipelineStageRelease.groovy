import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /** Can perform both to cloud foundry and neo targets. Preferred over cloudFoundryDeploy and neoDeploy, if configured. */
    'multicloudDeploy',
    /** For Cloud Foundry use-cases: Performs deployment to Cloud Foundry space/org. */
    'cloudFoundryDeploy',
    /** Performs health check in order to prove that deployment was successful. */
    'healthExecuteCheck',
    /** For Neo use-cases: Performs deployment to Neo landscape. */
    'neoDeploy',
    /** For TMS use-cases: Performs upload to Transport Management Service node*/
    'tmsUpload',
    /** Publishes release information to GitHub. */
    'githubPublishRelease',
    /** Executes smoke tests by running the npm script 'ci-smoke' defined in the project's package.json file. */
    'npmExecuteEndToEndTests'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * This stage is responsible to release/deploy artifacts into your productive landscape.<br />
 */
@GenerateStageDocumentation(defaultStageName = 'Release')
void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    def stageName = parameters.stageName?:env.STAGE_NAME

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('multicloudDeploy', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.multicloudDeploy)
        .addIfEmpty('cloudFoundryDeploy', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.cloudFoundryDeploy)
        .addIfEmpty('githubPublishRelease', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.githubPublishRelease)
        .addIfEmpty('healthExecuteCheck', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.healthExecuteCheck)
        .addIfEmpty('tmsUpload', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.tmsUpload)
        .addIfEmpty('neoDeploy', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.neoDeploy)
        .addIfEmpty('npmExecuteEndToEndTests', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.npmExecuteEndToEndTests)
        .use()

    piperStageWrapper (script: script, stageName: stageName) {

        // telemetry reporting
        utils.pushToSWA([step: STEP_NAME], config)

        // Prefer the newer multicloudDeploy step if it is configured as it is more capable
        if (config.multicloudDeploy) {
            durationMeasure(script: script, measurementName: 'deploy_release_multicloud_duration') {
                multicloudDeploy(script: script, stage: stageName)
            }
        } else {
            if (config.cloudFoundryDeploy) {
                durationMeasure(script: script, measurementName: 'deploy_release_cf_duration') {
                    cloudFoundryDeploy script: script
                }
            }

            if (config.neoDeploy) {
                durationMeasure(script: script, measurementName: 'deploy_release_neo_duration') {
                    neoDeploy script: script
                }
            }
        }

        if (config.tmsUpload) {
            durationMeasure(script: script, measurementName: 'upload_release_tms_duration') {
                tmsUpload script: script
            }
        }

        if (config.healthExecuteCheck) {
            healthExecuteCheck script: script
        }

        if (config.npmExecuteEndToEndTests) {
            durationMeasure(script: script, measurementName: 'npmExecuteEndToEndTests_duration') {
                npmExecuteEndToEndTests script: script, stageName: stageName, runScript: 'ci-smoke'
            }
        }

        if (config.githubPublishRelease) {
            githubPublishRelease script: script
        }

    }
}
