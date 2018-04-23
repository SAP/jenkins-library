package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

import com.sap.piper.MapUtils

class ConfigurationMerger {
    @NonCPS
    def static merge(Map configs, Set configKeys, Map defaults) {
        Map filteredConfig = configKeys?configs.subMap(configKeys):configs
        Map merged = [:]

        defaults = defaults ?: [:]

        merged.putAll(defaults)

        for(String key : filteredConfig.keySet())
            if(MapUtils.isMap(filteredConfig[key]))
                merged[key] = merge(filteredConfig[key], null, defaults[key])
            else if(filteredConfig[key] != null)
                merged[key] = filteredConfig[key]
            // else: keep defaults value and omit null values from config
        return merged
    }

    @NonCPS
    def static merge(
        Map parameters, Set parameterKeys,
        Map configuration, Set configurationKeys,
        Map defaults=[:]
    ){
        Map merged
        merged = merge(configuration, configurationKeys, defaults)
        merged = merge(parameters, parameterKeys, merged)
        return merged
    }

    @NonCPS
    def static merge(
        def script, def stepName,
        Map parameters, Set parameterKeys,
        Set stepConfigurationKeys
    ) {
          merge(script, stepName, parameters, parameterKeys, [:], stepConfigurationKeys)
    }

    @NonCPS
    def static merge(
        def script, def stepName,
        Map parameters, Set parameterKeys,
        Map pipelineData,
        Set stepConfigurationKeys
    ) {
        Map stepDefaults = ConfigurationLoader.defaultStepConfiguration(script, stepName)
        Map stepConfiguration = ConfigurationLoader.stepConfiguration(script, stepName)

        mergeWithPipelineData(parameters, parameterKeys, pipelineData, stepConfiguration, stepConfigurationKeys, stepDefaults)
    }

    @NonCPS
    def static mergeWithPipelineData(Map parameters, Set parameterKeys,
                            Map pipelineDataMap,
                            Map configurationMap, Set configurationKeys,
                            Map stepDefaults=[:]
    ){
        Map merged
        merged = merge(configurationMap, configurationKeys, stepDefaults)
        merged = merge(pipelineDataMap, null, merged)
        merged = merge(parameters, parameterKeys, merged)

        return merged
    }
}
