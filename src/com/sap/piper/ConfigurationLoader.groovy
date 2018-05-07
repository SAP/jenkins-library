package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

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
        return script?.commonPipelineEnvironment?.configuration?.general ?: [:]
    }

    @NonCPS
    static Map defaultGeneralConfiguration(script){
        return DefaultValueCache.getInstance()?.getDefaultValues()?.general ?: [:]
    }

    @NonCPS
    private static Map loadConfiguration(script, String type, String entryName, ConfigurationType configType){
        switch (configType) {
            case ConfigurationType.CUSTOM_CONFIGURATION:
                return script?.commonPipelineEnvironment?.configuration?.get(type)?.get(entryName) ?: [:]
            case ConfigurationType.DEFAULT_CONFIGURATION:
                return DefaultValueCache.getInstance()?.getDefaultValues()?.get(type)?.get(entryName) ?: [:]
            default:
                throw new IllegalArgumentException("Unknown configuration type: ${configType}")
        }
    }
}
