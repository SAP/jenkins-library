package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

import com.sap.piper.MapUtils

class ConfigurationMerger {
    @NonCPS
    def static merge(Map configs, List configKeys, Map defaults=[:]) {
        return merge(configs, MapUtils.fromList(configKeys), defaults)
    }

    @NonCPS
    def static merge(Map configs, Map configKeys, Map defaults = [:]) {
        Map merged = [:]
        merged.putAll(defaults)

        if(configs != null && configKeys){
            configs = configs.subMap(configKeys.keySet())
            for(String key : configKeys.keySet())
                if(MapUtils.isMap(configKeys[key]))
                    merged[key] = merge(configs[key], configKeys[key], defaults[key])
                else if(configs[key] != null)
                    merged[key] = configs[key]
        }
        return merged
    }

    @NonCPS
    def static merge(
        Map parameters, List parameterKeys,
        Map configuration, List configurationKeys,
        Map defaults=[:]
    ){
        return merge(
            parameters, MapUtils.fromList(parameterKeys),
            configuration, MapUtils.fromList(configurationKeys),
            defaults)
    }

    @NonCPS
    def static merge(
        Map parameters, Map parameterKeys,
        Map configuration, Map configurationKeys,
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
