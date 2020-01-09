import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /** For Cloud Foundry use-cases: Performs deployment to Cloud Foundry space/org. */
    'cloudFoundryDeploy',
    /** Performs behavior-driven tests using Gauge test framework against the deployed application/service. */
    'gaugeExecuteTests',
    /**
     * Performs health check in order to prove one aspect of operational readiness.
     * In order to be able to respond to health checks from infrastructure components (like load balancers) it is important to provide one unprotected application endpoint which allows a judgement about the health of your application.
     */
    'healthExecuteCheck',
    /** For Neo use-cases: Performs deployment to Neo landscape. */
    'neoDeploy',
    /** Performs API testing using Newman against the deployed application/service. */
    'newmanExecute',
    /** Publishes test results to Jenkins. It will automatically be active in cases tests are executed. */
    'testsPublishResults',
    /** Performs end-to-end UI testing using UIVeri5 test framework against the deployed application/service. */
    'uiVeri5ExecuteTests'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * In this stage the application/service is typically deployed and automated acceptance tests are executed.<br />
 * This is to make sure that
 *
 * * new functionality is tested end-to-end
 * * there is no end-to-end regression in existing functionality
 *
 */
@GenerateStageDocumentation(defaultStageName = "Acceptance")
void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    def stageName = parameters.stageName?:env.STAGE_NAME

    Map config = ConfigurationHelper.newInstance(this, script)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('cloudFoundryDeploy', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.cloudFoundryDeploy)
        .addIfEmpty('gaugeExecuteTests', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.gaugeExecuteTests)
        .addIfEmpty('healthExecuteCheck', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.healthExecuteCheck)
        .addIfEmpty('neoDeploy', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.neoDeploy)
        .addIfEmpty('newmanExecute', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.newmanExecute)
        .addIfEmpty('uiVeri5ExecuteTests', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.uiVeri5ExecuteTests)
        .use()

    piperStageWrapper (script: script, stageName: stageName) {

        // telemetry reporting
        utils.pushToSWA([step: STEP_NAME], config)


        if (config.cloudFoundryDeploy) {
            durationMeasure(script: script, measurementName: 'deploy_test_duration') {
                cloudFoundryDeploy script: script
            }
        }

        if (config.neoDeploy) {
            durationMeasure(script: script, measurementName: 'deploy_test_duration') {
                neoDeploy script: script
            }
        }

        if (config.healthExecuteCheck) {
            healthExecuteCheck script: script
        }


        def publishMap = [script: script]
        def publishResults = false

        if (config.gaugeExecuteTests) {
            durationMeasure(script: script, measurementName: 'gauge_duration') {
                publishResults = true
                gaugeExecuteTests script: script
                publishMap += [gauge: [archive: true]]
            }
        }

        if (config.newmanExecute) {
            durationMeasure(script: script, measurementName: 'newman_duration') {
                publishResults = true
                newmanExecute script: script
            }
        }

        if (config.uiVeri5ExecuteTests) {
            durationMeasure(script: script, measurementName: 'uiveri5_duration') {
                publishResults = true
                uiVeri5ExecuteTests script: script
            }
        }

        if (publishResults) {
            testsPublishResults publishMap
        }
    }
}
