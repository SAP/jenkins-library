def call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: 'setupCommonPipelineEnvironment', stepParameters: parameters) {

        def script = parameters.script

        prepareDefaultValues script: script

        String configFile = parameters.get('configFile')

        loadConfigurationFromFile(script, configFile)
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
    String defaultYmlConfigFile = 'pipeline_config.yml'

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
