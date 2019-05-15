package com.sap.piper.config

import static com.sap.piper.MapUtils.pruneNulls

class ConfigCache {

    private static ConfigCache INSTANCE = null
    private static PIPER_OS_DEFAULTS = 'default_pipeline_environment.yml'

    final Map piperDefaults  // resources/default_pipeline_env (immutable

    //final Map<String, Map> customDefaults = [:] // next step: custom layers (n), immutable

    //final Map projectConfig = [:] // .pipeline/config.yml immutable

    private ConfigCache(Script steps /*, Set customDefaults */ /*, String projectConfig = '.pipeline/config.yml' */) {

        if(steps == null) throw new NullPointerException('Steps not available.')
        // next step: make immutable
        piperDefaults = pruneNulls(steps.readYaml(text: steps.libraryResource(PIPER_OS_DEFAULTS)))

        steps.echo "Loading configuration file '${PIPER_OS_DEFAULTS}'"

        // next step: read customConfig

        // next step: read projec config
    }

    static synchronized ConfigCache getInstance(Script steps) {

        if(INSTANCE == null) {
            INSTANCE = new ConfigCache(steps)
        }
        INSTANCE
    }

    // next step: Think about reset strategy for testing. In productive coding there should be no reset method.
    public static void reset(){
        INSTANCE = null
    }

}
