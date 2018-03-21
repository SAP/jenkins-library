package com.sap.piper

class ConfigurationHelper implements Serializable {
    static def loadStepDefaults(Script step){
        return new ConfigurationHelper(step)
            .initDefaults(step)
            .loadDefaults()
    }

    private Map config
    private String name

    ConfigurationHelper(Script step){
        name = step.STEP_NAME
        if(!name) throw new IllegalArgumentException('Step has no public name property!')
    }

    private final ConfigurationHelper initDefaults(Script step){
        step.prepareDefaultValues()
        return this
    }

    private final ConfigurationHelper loadDefaults(){
        config = ConfigurationLoader.defaultStepConfiguration(null, name)
        return this
    }

    ConfigurationHelper mixinStepConfig(commonPipelineEnvironment, Set filter = null){
        if(!name) throw new IllegalArgumentException('Step has no public name property!')
        Map stepConfiguration = ConfigurationLoader.stepConfiguration([commonPipelineEnvironment: commonPipelineEnvironment], name)
        return mixin(stepConfiguration, filter)
    }

    ConfigurationHelper mixin(Map parameters, Set filter = null){
        config = ConfigurationMerger.merge(parameters, filter, config)
        return this
    }

    Map use(){ return config }

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
