package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

class ConfigurationMerger {
    @NonCPS
    static Map merge(Map configs, Set configKeys, Map defaults) {
        Map filteredConfig = configKeys?configs.subMap(configKeys):configs

        return MapUtils.merge(MapUtils.pruneNulls(defaults),
                              MapUtils.pruneNulls(filteredConfig))
    }

    @NonCPS
    static Map merge(
        Map parameters, Set parameterKeys,
        Map configuration, Set configurationKeys,
        Map defaults=[:]
    ){
        Map merged
        merged = merge(configuration, configurationKeys, defaults)
        merged = merge(parameters, parameterKeys, merged)
        return merged
    }
}
