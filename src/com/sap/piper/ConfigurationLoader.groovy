package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

@API(deprecated = true)
class ConfigurationLoader implements Serializable {
    @NonCPS
    static Map stepConfiguration(String stepName) {
        return loadConfiguration('steps', stepName, ConfigurationType.CUSTOM_CONFIGURATION)
    }

    @NonCPS
    static Map stageConfiguration(String stageName) {
        return loadConfiguration('stages', stageName, ConfigurationType.CUSTOM_CONFIGURATION)
    }

    @NonCPS
    static Map defaultStepConfiguration(String stepName) {
        return loadConfiguration('steps', stepName, ConfigurationType.DEFAULT_CONFIGURATION)
    }

    @NonCPS
    static Map defaultStageConfiguration(String stageName) {
        return loadConfiguration('stages', stageName, ConfigurationType.DEFAULT_CONFIGURATION)
    }

    @NonCPS
    static Map generalConfiguration(){
        try {
            return DefaultValueCache.getInstance().getProjectConfig()?.general ?: [:]
        } catch (groovy.lang.MissingPropertyException mpe) {
            return [:]
        }
    }

    @NonCPS
    static Map defaultGeneralConfiguration(){
        return DefaultValueCache.getInstance()?.getDefaultValues()?.general ?: [:]
    }

    @NonCPS
    static Map postActionConfiguration(String actionName){
        return loadConfiguration('postActions', actionName, ConfigurationType.CUSTOM_CONFIGURATION)
    }

    @NonCPS
    private static Map loadConfiguration(String type, String entryName, ConfigurationType configType){
        switch (configType) {
            case ConfigurationType.CUSTOM_CONFIGURATION:
                try {
                    return DefaultValueCache.getInstance()?.getProjectConfig()?.get(type)?.get(entryName) ?: [:]
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
