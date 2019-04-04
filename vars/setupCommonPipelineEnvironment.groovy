import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = ['collectTelemetryData']

void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters)

        prepareDefaultValues script: script, customDefaults: parameters.customDefaults

        String configFile = parameters.get('configFile')

        loadConfigurationFromFile(script, configFile)

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .use()

        (parameters.utils ?: new Utils()).pushToSWA([
            step: STEP_NAME,
            stepParamKey4: 'customDefaults',
            stepParam4: parameters.customDefaults?'true':'false'
        ], config)

        script.commonPipelineEnvironment.setInfluxStepData('build_url', env.BUILD_URL)
        script.commonPipelineEnvironment.setInfluxPipelineData('build_url', env.BUILD_URL)
    }
}

private loadConfigurationFromFile(script, String configFile) {

    String defaultYmlConfigFile = '.pipeline/config.yml'

    if (configFile) {
        script.commonPipelineEnvironment.configuration = readYaml(file: configFile)
    } else if (fileExists(defaultYmlConfigFile)) {
        script.commonPipelineEnvironment.configuration = readYaml(file: defaultYmlConfigFile)
    }
}
