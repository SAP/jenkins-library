def call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: 'setupCommonPipelineEnvironment', stepParameters: parameters) {

        prepareDefaultValues script: script

        String configFile = parameters.get('configFile')

        loadConfigurationFromFile(configFile)
    }
}

private boolean isYaml(String fileName) {
    return fileName.endsWith(".yml") || fileName.endsWith(".yaml")
}

private boolean isProperties(String fileName) {
    return fileName.endsWith(".properties")
}

private loadConfigurationFromFile(script, String configFile) {

    String defaultPropertiesYmlConfigFile = '.pipeline/config.properties'
    String defaultYmlConfigFile = 'pipeline_config.yml'

    if (configFile?.trim()?.length() > 0 && isProperties(configFile)) {
        Map configMap = readProperties(file: configFile)
        script.commonPipelineEnvironment.setConfigProperties(configMap)
    } else if (fileExists(defaultPropertiesYmlConfigFile)) {
        Map configMap = readProperties(file: defaultPropertiesYmlConfigFile)
        script.commonPipelineEnvironment.setConfigProperties(configMap)
    }

    if (configFile?.trim()?.length() > 0 && isYaml(configFile)) {
        script.commonPipelineEnvironment.configuration = readYaml(file: configFile)
    } else if (fileExists(defaultYmlConfigFile)) {
        script.commonPipelineEnvironment.configuration = readYaml(file: defaultYmlConfigFile)
    }
}
