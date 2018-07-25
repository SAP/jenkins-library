package com.sap.piper

class ConfigurationHelper implements Serializable {
    static def loadStepDefaults(Script step){
        return new ConfigurationHelper(step)
            .initDefaults(step)
            .loadDefaults()
    }

    private Map config = [:]
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
        config = ConfigurationLoader.defaultGeneralConfiguration()
        mixin(ConfigurationLoader.defaultStepConfiguration(null, name))
        return this
    }

    ConfigurationHelper mixinGeneralConfig(commonPipelineEnvironment, Set filter = null){
        Map stepConfiguration = ConfigurationLoader.generalConfiguration([commonPipelineEnvironment: commonPipelineEnvironment])
        return mixin(stepConfiguration, filter)
    }

    ConfigurationHelper mixinStageConfig(commonPipelineEnvironment, stageName, Set filter = null){
        Map stageConfiguration = ConfigurationLoader.stageConfiguration([commonPipelineEnvironment: commonPipelineEnvironment], stageName)
        return mixin(stageConfiguration, filter)
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

    ConfigurationHelper handleCompatibility(Script step, Map compatibleParameters){
        config = ConfigurationMerger.merge(recurseCompatibility(step, compatibleParameters, config), null, config)
        return this
    }

    private Map recurseCompatibility(Script step, Map compatibleParameters, String paramStructure = '', configMap) {
        Map newConfig = [:]
        compatibleParameters.each {entry ->
            if (entry.getValue() instanceof Map) {
                paramStructure = (paramStructure ? paramStructure + '.' : '') + entry.getKey()
                newConfig[entry.getKey()] = recurseCompatibility(step, entry.getValue(), paramStructure, (configMap!=null ? configMap[entry.getKey()] : null))
            } else {
                if (configMap == null || (configMap != null && configMap[entry.getKey()] == null)) {
                    newConfig[entry.getKey()] = config[entry.getValue()]
                    def paramName = (paramStructure ? paramStructure + '.' : '') + entry.getKey()
                    if (step) {
                        step.echo ("[INFO] The parameter '${entry.getValue()}' is COMPATIBLE to the parameter '${paramName}'")
                    }
                }
            }
        }
        return newConfig
    }

    Map dependingOn(dependentKey){
        return [
            mixin: {key ->
                def dependentValue = config[dependentKey]
                if(config[key] == null && dependentValue && config[dependentValue])
                    config[key] = config[dependentValue][key]
                return this
            }
        ]
    }

    ConfigurationHelper addIfEmpty(key, value){
        if (config[key] instanceof Boolean) {
            return this
        } else if (!config[key]){
            config[key] = value
        }
        return this
    }

    Map use(){ return config }

    ConfigurationHelper(Map config = [:]){
        this.config = config
    }

    def getConfigProperty(key) {
        return getConfigPropertyNested(config, key)
    }

    def getConfigProperty(key, defaultValue) {
        def value = getConfigProperty(key)
        if (value == null) {
            return defaultValue
        }
        return value
    }

    private getConfigPropertyNested(Map config, key) {

        def separator = '/'

        // reason for cast to CharSequence: String#tokenize(./.) causes a deprecation warning.
        List parts = (key in String) ? (key as CharSequence).tokenize(separator) : ([key] as List)

        if(config[parts.head()] != null) {

            if(config[parts.head()] in Map && ! parts.tail().isEmpty()) {
                return getConfigPropertyNested(config[parts.head()], (parts.tail() as Iterable).join(separator))
            }

            if (config[parts.head()].class == String) {
                return (config[parts.head()] as String).trim()
            }
        }


        return config[parts.head()]
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

    def getMandatoryProperty(key, defaultValue = null, errorMessage = null) {

        def paramValue = getConfigProperty(key, defaultValue)

        if (paramValue == null) {
            if(! errorMessage) errorMessage = "ERROR - NO VALUE AVAILABLE FOR ${key}"
            throw new IllegalArgumentException(errorMessage)
        }
        return paramValue
    }

    def withMandatoryProperty(key, errorMessage = null){
        getMandatoryProperty(key, null, errorMessage)
        return this
    }
}
