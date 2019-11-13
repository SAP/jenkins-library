package com.sap.piper

// script is present in the signatures in order to keep api compatibility.
// The script referenced is not used inside the method bodies.

@API(deprecated = true)
class ConfigurationLoader implements Serializable {

    static Map stepConfiguration(String stepName) {
        return loadConfiguration('steps', stepName, ConfigurationType.CUSTOM_CONFIGURATION)
    }
    @Deprecated
    /** Use stepConfiguration(stepName) instead */
    static Map stepConfiguration(script, String stepName) {
        return stepConfiguration(stepName)
    }

    static Map stageConfiguration(String stageName) {
        return loadConfiguration('stages', stageName, ConfigurationType.CUSTOM_CONFIGURATION)
    }
    @Deprecated
    /** Use stageConfiguration(stageName) instead */
    static Map stageConfiguration(script, String stageName) {
        return stageConfiguration(stageName)
    }

    static Map defaultStepConfiguration(String stepName) {
        return loadConfiguration('steps', stepName, ConfigurationType.DEFAULT_CONFIGURATION)
    }
    @Deprecated
    /** Use defaultStepConfiguration(stepName) instead */
    static Map defaultStepConfiguration(script, String stepName) {
        return defaultStepConfiguration(stepName)
    }

    static Map defaultStageConfiguration(String stageName) {
        return loadConfiguration('stages', stageName, ConfigurationType.DEFAULT_CONFIGURATION)
    }
    @Deprecated
    /** Use defaultStageConfiguration(stepName) instead */
    static Map defaultStageConfiguration(script, String stageName) {
        return defaultStageConfiguration(stageName)
    }

    static Map generalConfiguration(){
        try {
            return CommonPipelineEnvironment.getInstance()?.configuration?.general ?: [:]
        } catch (groovy.lang.MissingPropertyException mpe) {
            return [:]
        }
    }
    @Deprecated
    /** Use generalConfiguration() instead */
    static Map generalConfiguration(script){
        return generalConfiguration()
    }

    static Map defaultGeneralConfiguration(){
        return DefaultValueCache.getInstance()?.getDefaultValues()?.general ?: [:]
    }
    @Deprecated
    /** Use defaultGeneralConfiguration() instead */
    static Map defaultGeneralConfiguration(script){
        return defaultGeneralConfiguration()
    }

    static Map postActionConfiguration(String actionName){
        return loadConfiguration('postActions', actionName, ConfigurationType.CUSTOM_CONFIGURATION)
    }
    @Deprecated
    /** Use postActionConfiguration() instead */
    static Map postActionConfiguration(script, String actionName){
        return postActionConfiguration(actionName)
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
