package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

import com.sap.piper.MapUtils

class ConfigurationMerger {
    @NonCPS
    def static merge(Map configs, List configKeys, Map defaults=[:]) {
        Map merged = [:]
        merged.putAll(defaults)
        merged.putAll(filterByKeyAndNull(configs, configKeys))

        return merged
    }

    @NonCPS
    def static merge(Map configs, Map configKeys, Map defaults = [:]) {
        Map merged = [:]
        merged.putAll(defaults)
        if(configs != null)
            for(String key : defaults.keySet())
                if(MapUtils.isMap(defaults[key]))
                    merged[key] = merge(configs[key], configKeys[key], defaults[key])
                else
                    merged[key] = configs[key]
        return merged
    }

    @NonCPS
    def static merge(Map parameters, List parameterKeys, Map configurationMap, List configurationKeys, Map defaults=[:]){
        Map merged = merge(configurationMap, configurationKeys, defaults)
        merged.putAll(filterByKeyAndNull(parameters, parameterKeys))

        return merged
    }

    @NonCPS
    def static mergeDeepStructure(Map parameters, Map parameterKeys, Map configuration, Map configurationKeys, Map defaults=[:]){
        Map merged = [:]
        merged.putAll(defaults)
        merged = merge(configuration, configurationKeys, merged)
        merged = merge(parameters, parameterKeys, merged)
        return merged
    }

    @NonCPS
    def static mergeWithPipelineData(Map parameters, List parameterKeys,
                            Map pipelineDataMap,
                            Map configurationMap, List configurationKeys,
                            Map stepDefaults=[:]
    ){
        Map merged = [:]
        merged.putAll(stepDefaults)
        merged.putAll(filterByKeyAndNull(configurationMap, configurationKeys))
        merged.putAll(pipelineDataMap)
        merged.putAll(filterByKeyAndNull(parameters, parameterKeys))

        return merged
    }

    @NonCPS
    private static filterByKeyAndNull(Map map, List keys) {
        Map filteredMap = map.findAll {
            if(it.value == null){
                return false
            }
            return true
        }

        if(keys == null) {
            return filteredMap
        }

        return filteredMap.subMap(keys)
    }
}
