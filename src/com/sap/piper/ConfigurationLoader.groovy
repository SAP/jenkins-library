package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

@API(deprecated = true)
class ConfigurationLoader implements Serializable {
    @NonCPS
    static Map stepConfiguration(script, String stepName) {
        return loadConfiguration(script, 'steps', stepName, ConfigurationType.CUSTOM_CONFIGURATION)
    }

    @NonCPS
    static Map stageConfiguration(script, String stageName) {
        return loadConfiguration(script, 'stages', stageName, ConfigurationType.CUSTOM_CONFIGURATION)
    }

    @NonCPS
    static Map defaultStepConfiguration(script, String stepName) {
        return loadConfiguration(script, 'steps', stepName, ConfigurationType.DEFAULT_CONFIGURATION)
    }

    @NonCPS
    static Map defaultStageConfiguration(script, String stageName) {
        return loadConfiguration(script, 'stages', stageName, ConfigurationType.DEFAULT_CONFIGURATION)
    }

    @NonCPS
    static Map generalConfiguration(script){
        try {
            return script?.commonPipelineEnvironment?.configuration?.general ?: [:]
        } catch (groovy.lang.MissingPropertyException mpe) {
            return [:]
        }
    }

    @NonCPS
    static Map defaultGeneralConfiguration(script){
        return DefaultValueCache.getInstance()?.getDefaultValues()?.general ?: [:]
    }

    @NonCPS
    static Map postActionConfiguration(script, String actionName){
        return loadConfiguration(script, 'postActions', actionName, ConfigurationType.CUSTOM_CONFIGURATION)
    }

    @NonCPS
    private static Map loadConfiguration(script, String type, String entryName, ConfigurationType configType){
        switch (configType) {
            case ConfigurationType.CUSTOM_CONFIGURATION:
                try {
                    return script?.commonPipelineEnvironment?.configuration?.get(type)?.get(entryName) ?: [:]
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
