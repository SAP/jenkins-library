package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

class ConfigurationHelper implements Serializable {

    static ConfigurationHelper newInstance(Script step, Map config = [:]) {
        new ConfigurationHelper(step, config)
    }

    ConfigurationHelper loadStepDefaults() {
        this.step.prepareDefaultValues()
        this.config = ConfigurationLoader.defaultGeneralConfiguration()
        mixin(ConfigurationLoader.defaultStepConfiguration(null, name))
    }

    private Map config
    private Script step
    private String name
    private Map validationResults = null

    private ConfigurationHelper(Script step, Map config){
        this.config = config ?: [:]
        this.step = step
        this.name = step.STEP_NAME
        if(!this.name) throw new IllegalArgumentException('Step has no public name property!')
    }

    ConfigurationHelper collectValidationFailures() {
        validationResults = validationResults ?: [:]
        return this
    }

    ConfigurationHelper mixinGeneralConfig(commonPipelineEnvironment, Set filter = null, Map compatibleParameters = [:]){
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

    ConfigurationHelper addIfNull(key, value){
        if (config[key] == null){
            config[key] = value
        }
        return this
    }

    @NonCPS // required because we have a closure in the
            // method body that cannot be CPS transformed
    Map use(){
        handleValidationFailures()
        MapUtils.traverse(config, { v -> (v instanceof GString) ? v.toString() : v })
        if(config.verbose) step.echo "[${name}] Configuration: ${config}"
        return config
    }

    /* private */ def getConfigPropertyNested(key) {
        return getConfigPropertyNested(config, key)
    }

    /* private */ static getConfigPropertyNested(Map config, key) {

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

    @NonCPS
    private handleValidationFailures() {
        if(! validationResults) return
        if(validationResults.size() == 1) throw validationResults.values().first()
        String msg = 'ERROR - NO VALUE AVAILABLE FOR: ' +
            (validationResults.keySet().stream().collect() as Iterable).join(', ')
        IllegalArgumentException iae = new IllegalArgumentException(msg)
        validationResults.each { e -> iae.addSuppressed(e.value) }
        throw iae
    }

}
