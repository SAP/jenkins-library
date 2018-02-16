package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

import com.sap.piper.MapUtils

class ConfigurationMerger {
    @NonCPS
    def static merge(Map configs, List configKeys, Map defaults) {
        Map filteredConfig = configKeys?configs.subMap(configKeys):configs
        Map merged = [:]

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
        Map parameters, List parameterKeys,
        Map configuration, List configurationKeys,
        Map defaults=[:]
    ){
        Map merged
        merged = merge(configuration, configurationKeys, defaults)
        merged = merge(parameters, parameterKeys, merged)
        return merged
    }

    @NonCPS
    def static mergeWithPipelineData(Map parameters, List parameterKeys,
                            Map pipelineDataMap,
                            Map configurationMap, List configurationKeys,
                            Map stepDefaults=[:]
    ){
        Map merged
        merged = merge(configurationMap, configurationKeys, stepDefaults)
        merged = merge(pipelineDataMap, null, merged)
        merged = merge(parameters, parameterKeys, merged)

        return merged
    }

    @NonCPS
    def static merge(
        Map parameters, List parameterKeys,
        Map stepConfigurationMap, List stepConfigurationKeys, Map stepConfigurationDefaults,
        Map generalConfigurationMap, List generalConfigurationKeys, Map generalConfigurationDefaults=[:]
    ){
        Map merged
        merged = merge(stepConfigurationDefaults, null, generalConfigurationDefaults)
        merged = merge(generalConfigurationMap, generalConfigurationKeys, merged)
        merged = merge(stepConfigurationMap, stepConfigurationKeys, merged)
        merged = merge(parameters, parameterKeys, merged)

        return merged
    }
}
