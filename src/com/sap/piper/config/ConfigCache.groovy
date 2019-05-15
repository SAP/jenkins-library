package com.sap.piper

import static com.sap.piper.MapUtils.pruneNulls

class ConfigCache {

    private static ConfigCache INSTANCE = null

    Map piperDefaults = [:]  // resources/default_pipeline_env (immutable

    private ConfigCache() {
    }

    Map<String, Map<?,?>> customDefaults = [:] // custom layers (n), immutable

    Map projectConfig = [:] // .pipeline/config.yml immutable

    static synchronized ConfigCache getInstance(Script steps) {
        
        if(INSTANCE == null) {
            INSTANCE = new ConfigCache()
            INSTANCE.initialize(steps)
        }
        INSTANCE
    }
    
    void initialize(def steps /*, Set customDefaults */ /*, String projectConfig = '.pipeline/config.yml' */) {

        piperDefaults = pruneNulls(steps.readYaml(text: steps.libraryResource('default_pipeline_environment.yml'))) // make immutable
        //customDefaults = // read custome default
    }
}