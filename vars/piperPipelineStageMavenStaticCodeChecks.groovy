import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationLoader
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.QualityCheck
import com.sap.piper.ReportAggregator
import com.sap.piper.Utils

import static com.sap.piper.Prerequisites.checkScript

import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field STAGE_STEP_KEYS = [
    /** Executes static code checks for Maven based projects. The plugins SpotBugs and PMD are used. */
    'mavenExecuteStaticCodeChecks'
]
@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Execute static code checks for Maven based projects. This stage enforces SAP Cloud SDK specific PND rulesets as well as SpotBugs include filter.
 *
 */
@GenerateStageDocumentation(defaultStageName = 'mavenStaticCodeChecks')
void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: null

    def utils = parameters.juStabUtils ?: new Utils()

    def stageName = parameters.stageName?:env.STAGE_NAME

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('mavenExecuteStaticCodeChecks', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.mavenExecuteStaticCodeChecks)
        .use()

    piperStageWrapper(stageName: stageName, script: script) {

        // telemetry reporting
        utils.pushToSWA([step: STEP_NAME], config)

        if (config.mavenExecuteStaticCodeChecks) {
            mavenExecuteStaticCodeChecks(script: script)
        }
    }
}
