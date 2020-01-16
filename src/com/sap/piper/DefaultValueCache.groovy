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

    static getInstance(Binding scriptBinding){
        if (instance) {
            return instance
        }
        if(scriptBinding?.hasVariable("defaultValueCacheInstance")) {
            instance = scriptBinding?.getProperty("defaultValueCacheInstance")
            return instance
        }
    }

    static createInstance(Binding scriptBinding, Map defaultValues, List customDefaults = []){
        instance = new DefaultValueCache(defaultValues, customDefaults)
        if(!scriptBinding?.hasVariable("defaultValueCacheInstance")) {
            scriptBinding?.setProperty("defaultValueCacheInstance", instance)
        }
        return instance
    }

    static boolean hasInstance(){
        return instance!=null
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

    static void prepare(Script workflowScript, Map parameters = [:]) {
        if(parameters == null) parameters = [:]
        if(!DefaultValueCache.hasInstance() || parameters.customDefaults) {
            def defaultValues = [:]
            def configFileList = ['default_pipeline_environment.yml']
            def customDefaults = parameters.customDefaults

            if(customDefaults in String)
                customDefaults = [customDefaults]
            if(customDefaults in List)
                configFileList += customDefaults
            for (def configFileName : configFileList){
                if(configFileList.size() > 1) workflowScript.echo "Loading configuration file '${configFileName}'"
                def configuration = workflowScript.readYaml text: workflowScript.libraryResource(configFileName)
                defaultValues = MapUtils.merge(
                        MapUtils.pruneNulls(defaultValues),
                        MapUtils.pruneNulls(configuration))
            }
            DefaultValueCache.createInstance(workflowScript.getBinding(), defaultValues, customDefaults)
        }
    }

    @Override
    public String toString() {
        return "DefaultValueCache{" +
            "defaultValues=" + defaultValues +
            '}'
    }
}
