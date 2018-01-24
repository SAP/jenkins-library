package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

class ConfigurationMerger {
    @NonCPS
    def static merge(Map configs, List configKeys, Map defaults=[:]) {
        Map merged = [:]
        merged.putAll(defaults)
        merged.putAll(filterByKeyAndNull(configs, configKeys))

        return merged
    }

    @NonCPS
    def static merge(Map parameters, List parameterKeys, Map configurationMap, List configurationKeys, Map defaults=[:]){
        Map merged = merge(configurationMap, configurationKeys, defaults)
        merged.putAll(filterByKeyAndNull(parameters, parameterKeys))

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
