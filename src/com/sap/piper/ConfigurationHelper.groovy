package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

@API
class ConfigurationHelper implements Serializable {

    def static SEPARATOR = '/'

    static ConfigurationHelper newInstance(Script step, Map config = [:]) {
        new ConfigurationHelper(step, config)
    }

    ConfigurationHelper loadStepDefaults(Map compatibleParameters = [:]) {
        DefaultValueCache.prepare(step)
        this.config = ConfigurationLoader.defaultGeneralConfiguration()
        mixin(ConfigurationLoader.defaultGeneralConfiguration(), null, compatibleParameters)
        mixin(ConfigurationLoader.defaultStepConfiguration(null, name), null, compatibleParameters)
    }

    private Map config
    private Script step
    private String name
    private Map validationResults = null
    private String dependingOn

    private ConfigurationHelper(Script step, Map config){
        this.config = config ?: [:]
        this.step = step
        this.name = step.STEP_NAME
        if(!this.name) throw new IllegalArgumentException('Step has no public name property!')
    }

    /*
     * By default this methods does nothing. With this method we are able to ensure that we do not call the
     * deprecated methods. Might be usefull during local development.
     */
    private static handleDeprecation(script, String methodName) {
        if(script != null) {
            def msg = "ConfigurationHelper.${methodName} was called with a script reference." +
                'This method is deprecated. Use the same method without the script reference'
            if(Boolean.getBoolean('com.sap.piper.failOnScriptReferenceInConfigurationHelper'))
                throw new RuntimeException(msg)
            if(Boolean.getBoolean('com.sap.piper.emitWarningOnScriptReferenceInConfigurationHelper') &&
                script instanceof Script) script.echo("[WARNING] ${msg}")
        }
    }


    ConfigurationHelper collectValidationFailures() {
        validationResults = validationResults ?: [:]
        return this
    }

    ConfigurationHelper mixinGeneralConfig(Set filter = null, Map compatibleParameters = [:]){
        mixinGeneralConfig(null, filter, compatibleParameters)
    }
    @Deprecated
    /** Use  mixinGeneralConfig without commonPipelineEnvironment*/
    ConfigurationHelper mixinGeneralConfig(commonPipelineEnvironment, Set filter = null, Map compatibleParameters = [:]){
        handleDeprecation(commonPipelineEnvironment, 'mixinGeneralConfig')
        Map generalConfiguration = ConfigurationLoader.generalConfiguration()
        return mixin(generalConfiguration, filter, compatibleParameters)
    }

    ConfigurationHelper mixinStageConfig(stageName, Set filter = null, Map compatibleParameters = [:]){
        mixinStageConfig(null, stageName, filter, compatibleParameters)
    }
    @Deprecated
    ConfigurationHelper mixinStageConfig(commonPipelineEnvironment, stageName, Set filter = null, Map compatibleParameters = [:]){
        handleDeprecation(commonPipelineEnvironment, 'mixinStageConfig')
        Map stageConfiguration = ConfigurationLoader.stageConfiguration(stageName)
        return mixin(stageConfiguration, filter, compatibleParameters)
    }

    ConfigurationHelper mixinStepConfig(Set filter = null, Map compatibleParameters = [:]){
        mixinStepConfig(null, filter, compatibleParameters)
    }
    @Deprecated
    ConfigurationHelper mixinStepConfig(commonPipelineEnvironment, Set filter = null, Map compatibleParameters = [:]){
        handleDeprecation(commonPipelineEnvironment, 'mixinStepConfig')
        Map stepConfiguration = ConfigurationLoader.stepConfiguration(name)
        return mixin(stepConfiguration, filter, compatibleParameters)
    }

    final ConfigurationHelper mixin(Map parameters, Set filter = null, Map compatibleParameters = [:]){
        if (parameters.size() > 0 && compatibleParameters.size() > 0) {
            parameters = ConfigurationMerger.merge(handleCompatibility(compatibleParameters, parameters), null, parameters)
        }
        if (filter) {
            filter.add('collectTelemetryData')
        }
        config = ConfigurationMerger.merge(parameters, filter, config)
        return this
    }

    private Map handleCompatibility(Map compatibleParameters, String paramStructure = '', Map configMap, Map newConfigMap = [:] ) {
        Map newConfig = [:]
        compatibleParameters.each {entry ->
            if (entry.getValue() instanceof Map) {
                def internalParamStructure = (paramStructure ? paramStructure + '.' : '') + entry.getKey()
                newConfig[entry.getKey()] = handleCompatibility(entry.getValue(), internalParamStructure, configMap, newConfig)
            } else {
                def configSubMap = configMap
                for(String key in paramStructure.tokenize('.')){
                    configSubMap = configSubMap?.get(key)
                }
                if (configSubMap == null || (configSubMap != null && configSubMap[entry.getKey()] == null)) {
                    def value = configMap[entry.getValue()]
                    if(null == value)
                        value = newConfigMap[entry.getValue()]
                    if (value != null) {
                        newConfig[entry.getKey()] = value
                        def paramName = (paramStructure ? paramStructure + '.' : '') + entry.getKey()
                        this.step.echo ("[INFO] The parameter '${entry.getValue()}' is COMPATIBLE to the parameter '${paramName}'")
                    }
                }
            }
        }
        return newConfig
    }

    ConfigurationHelper mixin(String key){
        def parts = tokenizeKey(key)
        def targetMap = config
        if(parts.size() > 1) {
            key = parts.last()
            parts.remove(key)
            targetMap = getConfigPropertyNested(config, parts.join(SEPARATOR))
        }
        def dependentValue = config[dependingOn]
        if(targetMap[key] == null && dependentValue && config[dependentValue])
            targetMap[key] = config[dependentValue][key]

        dependingOn = null
        return this
    }

    ConfigurationHelper dependingOn(dependentKey){
        dependingOn = dependentKey
        return this
    }

    ConfigurationHelper addIfEmpty(key, value){
        if (config[key] instanceof Boolean) {
            return this
        } else if (!config[key]){
            config[key] = value
        }
        return this
    }

    ConfigurationHelper addIfNull(key, value){
        if (config[key] == null){
            config[key] = value
        }
        return this
    }

    Map use(){
        handleValidationFailures()
        MapUtils.traverse(config, { v -> (v instanceof GString) ? v.toString() : v })
        if(config.verbose) step.echo "[${name}] Configuration: ${config}"
        return MapUtils.deepCopy(config)
    }

    /* private */ def getConfigPropertyNested(key) {
        return getConfigPropertyNested(config, key)
    }

    /* private */ static getConfigPropertyNested(Map config, key) {

        List parts = tokenizeKey(key)

        if (config[parts.head()] != null) {

            if (config[parts.head()] in Map && !parts.tail().isEmpty()) {
                return getConfigPropertyNested(config[parts.head()], parts.tail().join(SEPARATOR))
            }

            if (config[parts.head()].class == String) {
                return (config[parts.head()] as String).trim()
            }
        }
        return config[parts.head()]
    }

    /* private */  static tokenizeKey(String key) {
        // reason for cast to CharSequence: String#tokenize(./.) causes a deprecation warning.
        List parts = (key in String) ? (key as CharSequence).tokenize(SEPARATOR) : ([key] as List)
        return parts
    }

    private void existsMandatoryProperty(key, errorMessage) {

        def paramValue = getConfigPropertyNested(config, key)

        if (paramValue == null) {
            if(! errorMessage) errorMessage = "ERROR - NO VALUE AVAILABLE FOR ${key}"

            def iae = new IllegalArgumentException(errorMessage)
            if(validationResults == null) {
                throw iae
            }
            validationResults.put(key, iae)
        }
    }

    ConfigurationHelper withMandatoryProperty(key, errorMessage = null, condition = null){
        if(condition){
            if(condition(this.config))
                existsMandatoryProperty(key, errorMessage)
        }else{
            existsMandatoryProperty(key, errorMessage)
        }
        return this
    }

    ConfigurationHelper withPropertyInValues(String key, Set values){
        withMandatoryProperty(key)
        def value = config[key] instanceof GString ? config[key].toString() : config[key]
        if(! (value in values) ) {
            throw new IllegalArgumentException("Invalid ${key} = '${value}'. Valid '${key}' values are: ${values}.")
        }
        return this
    }

    private handleValidationFailures() {
        if(! validationResults) return
        if(validationResults.size() == 1) throw validationResults.values().first()
        String msg = 'ERROR - NO VALUE AVAILABLE FOR: ' + validationResults.keySet().join(', ')
        IllegalArgumentException iae = new IllegalArgumentException(msg)
        validationResults.each { e -> iae.addSuppressed(e.value) }
        throw iae
    }
}
