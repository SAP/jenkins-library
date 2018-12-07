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

        (parameters.utils ?: new Utils())
                   .pushToSWA([step: STEP_NAME, stepParam4: parameters.customDefaults?'true':'false',
                                                stepParam5: Boolean.toString( ! (script?.commonPipelineEnvironment?.getConfigProperties() ?: [:]).isEmpty())], config)
    }
}

private boolean isYaml(String fileName) {
    return fileName.endsWith(".yml") || fileName.endsWith(".yaml")
}

private boolean isProperties(String fileName) {
    return fileName.endsWith(".properties")
}

private loadConfigurationFromFile(script, String configFile) {

    String defaultPropertiesConfigFile = '.pipeline/config.properties'
    String defaultYmlConfigFile = '.pipeline/config.yml'

    if (configFile?.trim()?.length() > 0 && isProperties(configFile)) {
        Map configMap = readProperties(file: configFile)
        script.commonPipelineEnvironment.setConfigProperties(configMap)
    } else if (fileExists(defaultPropertiesConfigFile)) {
        Map configMap = readProperties(file: defaultPropertiesConfigFile)
        script.commonPipelineEnvironment.setConfigProperties(configMap)
    }

    if (configFile?.trim()?.length() > 0 && isYaml(configFile)) {
        script.commonPipelineEnvironment.configuration = readYaml(file: configFile)
    } else if (fileExists(defaultYmlConfigFile)) {
        script.commonPipelineEnvironment.configuration = readYaml(file: defaultYmlConfigFile)
    }
}
