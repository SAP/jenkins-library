package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.MapUtils

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

    @NonCPS
    static Map merge(
        def script, def stepName,
        Map parameters, Set parameterKeys,
        Set stepConfigurationKeys
    ) {
          merge(script, stepName, parameters, parameterKeys, [:], stepConfigurationKeys)
    }
}
