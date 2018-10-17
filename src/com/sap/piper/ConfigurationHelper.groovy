package com.sap.piper

class ConfigurationHelper implements Serializable {

    ConfigurationHelper loadStepDefaults() {
            initDefaults()
            loadDefaults()
    }

    private Map config
    private Script step
    private String name

    static ConfigurationHelper newInstance(Script step, Map config = [:]) {
        new ConfigurationHelper(step, config)
    }

    private ConfigurationHelper(Script step, Map config){
        this.config = config ?: [:]
        this.step = step
        this.name = step.STEP_NAME
        if(!this.name) throw new IllegalArgumentException('Step has no public name property!')
    }

    private final ConfigurationHelper initDefaults(){
        this.step.prepareDefaultValues()
        return this
    }

    private final ConfigurationHelper loadDefaults(){
        config = ConfigurationLoader.defaultGeneralConfiguration()
        mixin(ConfigurationLoader.defaultStepConfiguration(null, name))
        return this
    }

    ConfigurationHelper mixinGeneralConfig(commonPipelineEnvironment, Set filter = null,Map compatibleParameters = [:]){
        Map stepConfiguration = ConfigurationLoader.generalConfiguration([commonPipelineEnvironment: commonPipelineEnvironment])
        return mixin(stepConfiguration, filter, compatibleParameters)
    }

    ConfigurationHelper mixinStageConfig(commonPipelineEnvironment, stageName, Set filter = null, Map compatibleParameters = [:]){
        Map stageConfiguration = ConfigurationLoader.stageConfiguration([commonPipelineEnvironment: commonPipelineEnvironment], stageName)
        return mixin(stageConfiguration, filter, compatibleParameters)
    }

    ConfigurationHelper mixinStepConfig(commonPipelineEnvironment, Set filter = null, Map compatibleParameters = [:]){
        Map stepConfiguration = ConfigurationLoader.stepConfiguration([commonPipelineEnvironment: commonPipelineEnvironment], name)
        return mixin(stepConfiguration, filter, compatibleParameters)
    }

    ConfigurationHelper mixin(Map parameters, Set filter = null, Map compatibleParameters = [:]){
        if (parameters.size() > 0 && compatibleParameters.size() > 0) {
            parameters = ConfigurationMerger.merge(handleCompatibility(compatibleParameters, parameters), null, parameters)
        }
        if (filter) {
            filter.add('collectTelemetryData')
        }
        config = ConfigurationMerger.merge(parameters, filter, config)
        return this
    }

    private Map handleCompatibility(Map compatibleParameters, String paramStructure = '', Map configMap ) {
        Map newConfig = [:]
        compatibleParameters.each {entry ->
            if (entry.getValue() instanceof Map) {
                paramStructure = (paramStructure ? paramStructure + '.' : '') + entry.getKey()
                newConfig[entry.getKey()] = handleCompatibility(entry.getValue(), paramStructure, configMap)
            } else {
                def configSubMap = configMap
                for(String key in paramStructure.tokenize('.')){
                    configSubMap = configSubMap?.get(key)
                }
                if (configSubMap == null || (configSubMap != null && configSubMap[entry.getKey()] == null)) {
                    newConfig[entry.getKey()] = configMap[entry.getValue()]
                    def paramName = (paramStructure ? paramStructure + '.' : '') + entry.getKey()
                    if (configMap[entry.getValue()] != null) {
                        this.step.echo ("[INFO] The parameter '${entry.getValue()}' is COMPATIBLE to the parameter '${paramName}'")
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

    def withMandatoryProperty(key, errorMessage = null, condition = null){
        if(condition){
            if(condition(this.config))
                getMandatoryProperty(key, null, errorMessage)
        }else{
            getMandatoryProperty(key, null, errorMessage)
        }
        return this
    }
}
