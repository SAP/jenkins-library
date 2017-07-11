def call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: 'setupCommonPipelineEnvironment', stepParameters: parameters) {

        def configFile = parameters.get('configFile', '.pipeline/config.properties')
        def script = parameters.script

        Map configMap = [:]
        if (configFile.length() > 0)
            configMap = readProperties (file: configFile)
        script.commonPipelineEnvironment.setConfigProperties(configMap)

    }
}
