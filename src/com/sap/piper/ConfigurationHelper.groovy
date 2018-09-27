package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

class ConfigurationHelper implements Serializable {
    static ConfigurationHelper loadStepDefaults(Script step){
        return new ConfigurationHelper(step)
            .initDefaults(step)
            .loadDefaults()
    }

    private Map config = [:]
    private String name

    private Map validationResults = null

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

    ConfigurationHelper collectValidationFailures() {
        validationResults = validationResults ?: [:]
        return this
    }

    ConfigurationHelper mixinGeneralConfig(commonPipelineEnvironment, Set filter = null, Script step = null, Map compatibleParameters = [:]){
        Map stepConfiguration = ConfigurationLoader.generalConfiguration([commonPipelineEnvironment: commonPipelineEnvironment])
        return mixin(stepConfiguration, filter, step, compatibleParameters)
    }

    ConfigurationHelper mixinStageConfig(commonPipelineEnvironment, stageName, Set filter = null, Script step = null, Map compatibleParameters = [:]){
        Map stageConfiguration = ConfigurationLoader.stageConfiguration([commonPipelineEnvironment: commonPipelineEnvironment], stageName)
        return mixin(stageConfiguration, filter, step, compatibleParameters)
    }

    ConfigurationHelper mixinStepConfig(commonPipelineEnvironment, Set filter = null, Script step = null, Map compatibleParameters = [:]){
        if(!name) throw new IllegalArgumentException('Step has no public name property!')
        Map stepConfiguration = ConfigurationLoader.stepConfiguration([commonPipelineEnvironment: commonPipelineEnvironment], name)
        return mixin(stepConfiguration, filter, step, compatibleParameters)
    }

    ConfigurationHelper mixin(Map parameters, Set filter = null, Script step = null, Map compatibleParameters = [:]){
        if (parameters.size() > 0 && compatibleParameters.size() > 0) {
            parameters = ConfigurationMerger.merge(handleCompatibility(step, compatibleParameters, parameters), null, parameters)
        }
        if (filter) {
            filter.add('collectTelemetryData')
        }
        config = ConfigurationMerger.merge(parameters, filter, config)
        return this
    }

    private Map handleCompatibility(Script step, Map compatibleParameters, String paramStructure = '', Map configMap ) {
        Map newConfig = [:]
        compatibleParameters.each {entry ->
            if (entry.getValue() instanceof Map) {
                paramStructure = (paramStructure ? paramStructure + '.' : '') + entry.getKey()
                newConfig[entry.getKey()] = handleCompatibility(step, entry.getValue(), paramStructure, configMap)
            } else {
                def configSubMap = configMap
                for(String key in paramStructure.tokenize('.')){
                    configSubMap = configSubMap?.get(key)
                }
                if (configSubMap == null || (configSubMap != null && configSubMap[entry.getKey()] == null)) {
                    newConfig[entry.getKey()] = configMap[entry.getValue()]
                    def paramName = (paramStructure ? paramStructure + '.' : '') + entry.getKey()
                    if (step && configMap[entry.getValue()] != null) {
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

    @NonCPS // required because we have a closure in the
            // method body that cannot be CPS transformed
    Map use(){
        handleValidationFailures()
        MapUtils.traverse(config, { v -> (v instanceof GString) ? v.toString() : v })
        return config
    }

    ConfigurationHelper(Map config = [:]){
        this.config = config
    }

    /* private */ def getConfigPropertyNested(CharSequence key) {
        return getConfigPropertyNested(config, key)
    }

    /* private */ static getConfigPropertyNested(Map config, CharSequence key) {

        def separator = '/'

        List parts = key.tokenize(separator)

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

     private void existsMandatoryProperty(CharSequence key, errorMessage) {

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

    ConfigurationHelper withMandatoryProperty(CharSequence key, errorMessage = null, condition = null){
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
