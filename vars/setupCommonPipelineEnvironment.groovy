import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import com.sap.piper.analytics.InfluxData

import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /** */
    'collectTelemetryData'
]

@Field Set STEP_CONFIG_KEYS = []

@Field Set PARAMETER_KEYS = [
    /** Property file defining project specific settings.*/
    'configFile'
]

/**
 * Initializes the [`commonPipelineEnvironment`](commonPipelineEnvironment.md), which is used throughout the complete pipeline.
 *
 * !!! tip
 *    This step needs to run at the beginning of a pipeline right after the SCM checkout.
 *    Then subsequent pipeline steps consume the information from `commonPipelineEnvironment`; it does not need to be passed to pipeline steps explicitly.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {

    def script = checkScript(this, parameters)
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, script: script) {


        prepareDefaultValues script: script, customDefaults: parameters.customDefaults

        String configFile = parameters.get('configFile')

        loadConfigurationFromFile(script, configFile)

        Map config = ConfigurationHelper.newInstance(this, script)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .use()

        (parameters.utils ?: new Utils()).pushToSWA([
            step: STEP_NAME,
            stepParamKey4: 'customDefaults',
            stepParam4: parameters.customDefaults?'true':'false'
        ], config)

        InfluxData.addField('step_data', 'build_url', env.BUILD_URL)
        InfluxData.addField('pipeline_data', 'build_url', env.BUILD_URL)
    }
}

private loadConfigurationFromFile(script, String configFile) {

    String defaultYmlConfigFile = '.pipeline/config.yml'
    String defaultYamlConfigFile = '.pipeline/config.yaml'

    if (configFile) {
        script.commonPipelineEnvironment.configuration = readYaml(file: configFile)
    } else if (fileExists(defaultYmlConfigFile)) {
        script.commonPipelineEnvironment.configuration = readYaml(file: defaultYmlConfigFile)
    } else if (fileExists(defaultYamlConfigFile)) {
        script.commonPipelineEnvironment.configuration = readYaml(file: defaultYamlConfigFile)
    }
}
