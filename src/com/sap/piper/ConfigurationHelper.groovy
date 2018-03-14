package com.sap.piper

class ConfigurationHelper implements Serializable {
    static class ConfigLoader implements Serializable {
        private Map config
        private String name

        ConfigLoader(Script step){
            if(!step.STEP_NAME)
                throw new IllegalArgumentException('Step has no public name property!')
            name = step.STEP_NAME
            step.prepareDefaultValues()
            config = ConfigurationLoader.defaultStepConfiguration(step, name)
        }

        ConfigLoader mixinStepConfig(commonPipelineEnvironment, Set filter = null){
            Map stepConfiguration = ConfigurationLoader.stepConfiguration([commonPipelineEnvironment: commonPipelineEnvironment], name)
            return mixin(stepConfiguration, filter)
        }

        ConfigLoader mixin(Map parameters, Set filter = null){
            config = ConfigurationMerger.merge(parameters, filter, config)
            return this
        }

        Map use(){ return config }
    }

    static def loadStepDefaults(Script step){
        return new ConfigLoader(step)
    }

    private final Map config

    ConfigurationHelper(Map config = [:]){
        this.config = config
    }

    def getConfigProperty(key) {
        if (config[key] != null && config[key].class == String) {
            return config[key].trim()
        }
        return config[key]
    }

    def getConfigProperty(key, defaultValue) {
        def value = getConfigProperty(key)
        if (value == null) {
            return defaultValue
        }
        return value
    }

    def isPropertyDefined(key){

        def value = getConfigProperty(key)

        if(value == null){
            return false
        }

        if(value.class == String){
            return value?.isEmpty() == false
        }

        if(value){
            return true
        }

        return false
    }

    def getMandatoryProperty(key, defaultValue) {

        def paramValue = config[key]

        if (paramValue == null)
            paramValue = defaultValue

        if (paramValue == null)
            throw new Exception("ERROR - NO VALUE AVAILABLE FOR ${key}")
        return paramValue
    }

    def getMandatoryProperty(key) {
        return getMandatoryProperty(key, null)
    }
}
