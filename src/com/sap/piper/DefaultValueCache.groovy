package com.sap.piper

import com.sap.piper.MapUtils

@API
class DefaultValueCache implements Serializable {
    private static DefaultValueCache instance

    private Map defaultValues

    private List customDefaults = []

    private DefaultValueCache(Map defaultValues, List customDefaults){
        this.defaultValues = defaultValues
        if(customDefaults) {
            this.customDefaults.addAll(customDefaults)
        }
    }

    static getInstance(){
        return instance
    }

    static createInstance(Map defaultValues, List customDefaults = []){
        instance = new DefaultValueCache(defaultValues, customDefaults)
    }

    Map getDefaultValues(){
        return defaultValues
    }

    static reset(){
        instance = null
    }

    List getCustomDefaults() {
        def result = []
        result.addAll(customDefaults)
        return result
    }

    static void prepare(Script steps, Map parameters = [:]) {
        if(parameters == null) parameters = [:]
        if(!DefaultValueCache.getInstance() || parameters.customDefaults) {
            def defaultValues = [:]
            List paramCustomDefaults
            int numCustomDefaultsInConfig = 0

            if (parameters.customDefaults){
                paramCustomDefaults = parameters.customDefaults
                if (parameters.numCustomDefaultsInConfig){
                    numCustomDefaultsInConfig = parameters.numCustomDefaultsInConfig
                }
            } else {
                paramCustomDefaults = ['default_pipeline_environment.yml']
                steps.writeFile file: ".pipeline/${paramCustomDefaults[0]}", text: steps.libraryResource(paramCustomDefaults[0])
            }

            List customDefaults = []

            for (int i = 0; i < paramCustomDefaults.size(); i++) {
                if(paramCustomDefaults.size() > 1) steps.echo "Loading configuration file '${paramCustomDefaults[i]}'"
                def configuration = steps.readYaml file: ".pipeline/${paramCustomDefaults[i]}"

                // Only customDefaults not coming from project config are saved in customDefaults list to not have duplicated customDefaults when getConfig Go step is executed.
                // Since, the go step considers the customDefaults defined in project config in addition to the via CLI provided list of customDefaults.
                if (i <= paramCustomDefaults.size()-1-numCustomDefaultsInConfig){
                    customDefaults.add(paramCustomDefaults[i])
                }
                defaultValues = MapUtils.merge(
                    MapUtils.pruneNulls(defaultValues),
                    MapUtils.pruneNulls(configuration))
            }
            DefaultValueCache.createInstance(defaultValues, customDefaults)
        }
    }
}
