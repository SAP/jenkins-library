package com.sap.piper

// script is present in the signatures in order to keep api compatibility.
// The script referenced is not used inside the method bodies.

@API(deprecated = true)
class ConfigurationLoader implements Serializable {

    static Map stepConfiguration(String stepName) {
        return stepConfiguration(null, stepName)
    }
    @Deprecated
    /** Use stepConfiguration(stepName) instead */
    static Map stepConfiguration(script, String stepName) {
        return loadConfiguration('steps', stepName, ConfigurationType.CUSTOM_CONFIGURATION)
    }

    /*
     * By default this methods does nothing. With this method we are able to ensure that we do not call the
     * deprecated methods. Might be usefull during local development.
     */
    private static handleDeprecation(script, String methodName) {
        if(script != null) {
            def msg = "ConfigurationLoader.${methodName} was called with a script reference." +
                'This method is deprecated. Use the same method without the script reference'
            if(Boolean.getBoolean('com.sap.piper.failOnScriptReferenceInConfigurationLoader'))
                throw new RuntimeException(msg)
            if(Boolean.getBoolean('com.sap.piper.emitWarningOnScriptReferenceInConfigurationLoader') &&
                script instanceof Script) script.echo("[WARNING] ${msg}")
        }
    }

    static Map stageConfiguration(String stageName) {
        stageConfiguration(null, stageName)
    }
    @Deprecated
    /** Use stageConfiguration(stageName) instead */
    static Map stageConfiguration(script, String stageName) {
        handleDeprecation(script, 'stageConfiguration')
        return loadConfiguration('stages', stageName, ConfigurationType.CUSTOM_CONFIGURATION)
    }

    static Map defaultStepConfiguration(String stepName) {
        defaultStepConfiguration(null, stepName)
    }
    @Deprecated
    /** Use defaultStepConfiguration(stepName) instead */
    static Map defaultStepConfiguration(script, String stepName) {
        handleDeprecation(script, 'defaultStepConfiguration')
        return loadConfiguration('steps', stepName, ConfigurationType.DEFAULT_CONFIGURATION)
    }

    static Map defaultStageConfiguration(String stageName) {
        defaultStageConfiguration(null, stageName)
    }
    @Deprecated
    /** Use defaultStageConfiguration(stepName) instead */
    static Map defaultStageConfiguration(script, String stageName) {
        handleDeprecation(script, 'defaultStageConfiguration')
        return loadConfiguration('stages', stageName, ConfigurationType.DEFAULT_CONFIGURATION)
    }

    static Map generalConfiguration(){
        generalConfiguration(null)
    }
    @Deprecated
    /** Use generalConfiguration() instead */
    static Map generalConfiguration(script){
        handleDeprecation(script, 'generalConfiguration')
        try {
            return CommonPipelineEnvironment.getInstance()?.configuration?.general ?: [:]
        } catch (groovy.lang.MissingPropertyException mpe) {
            return [:]
        }
    }

    static Map defaultGeneralConfiguration(){
        defaultGeneralConfiguration(null)
    }
    @Deprecated
    /** Use defaultGeneralConfiguration() instead */
    static Map defaultGeneralConfiguration(script){
        handleDeprecation(script, 'defaultGeneralConfiguration')
        return DefaultValueCache.getInstance()?.getDefaultValues()?.general ?: [:]
    }

    static Map postActionConfiguration(String actionName){
        postActionConfiguration(null, actionName)
    }
    @Deprecated
    /** Use postActionConfiguration() instead */
    static Map postActionConfiguration(script, String actionName){
        handleDeprecation(script, 'postActionConfiguration')
        return loadConfiguration('postActions', actionName, ConfigurationType.CUSTOM_CONFIGURATION)
    }

    private static Map loadConfiguration(String type, String entryName, ConfigurationType configType){
        switch (configType) {
            case ConfigurationType.CUSTOM_CONFIGURATION:
                try {
                    return CommonPipelineEnvironment.getInstance()?.configuration?.get(type)?.get(entryName) ?: [:]
                } catch (groovy.lang.MissingPropertyException mpe) {
                    return [:]
                }

            case ConfigurationType.DEFAULT_CONFIGURATION:
                return DefaultValueCache.getInstance()?.getDefaultValues()?.get(type)?.get(entryName) ?: [:]
            default:
                throw new IllegalArgumentException("Unknown configuration type: ${configType}")
        }
    }
}
