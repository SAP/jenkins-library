package com.sap.piper.config

import static com.sap.piper.MapUtils.pruneNulls

class ConfigCache {

    private static ConfigCache INSTANCE = null
    private static PIPER_OS_DEFAULTS = 'default_pipeline_environment.yml'

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

    // TODO: Think about reset strategy for testing. In productive coding there should be no reset method.
    public static void reset(){
        INSTANCE = null
    }

    void initialize(def steps /*, Set customDefaults */ /*, String projectConfig = '.pipeline/config.yml' */) {

        piperDefaults = pruneNulls(steps.readYaml(text: steps.libraryResource(PIPER_OS_DEFAULTS))) // make immutable
        steps.echo "Loading configuration file '${PIPER_OS_DEFAULTS}'"

        //customDefaults = // read custome default

        // projectConfig = read config from '.pipeline/config.yml'
    }
}
