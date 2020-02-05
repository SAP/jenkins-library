package com.sap.piper

@API
class DefaultValueCache implements Serializable {
    private static DefaultValueCache instance

    private static final def defaultValuesRelativePath = '.pipeline/defaultValueCache/defaultValues'
    private static final def customDefaultsRelativePath = '.pipeline/defaultValueCache/customDefaults'
    private Map defaultValues

    private List customDefaults = []

    private DefaultValueCache(Map defaultValues, List customDefaults){
        this.defaultValues = defaultValues
        if(customDefaults) {
            this.customDefaults.addAll(customDefaults)
        }
    }

    static getInstance(script){
        if (!instance && script) {
            createInstanceFromPersistence(script)
        }
        return instance
    }

    static boolean hasInstance(){
        return instance!=null
    }

    static createInstance(Map defaultValues, List customDefaults = []){
        instance = new DefaultValueCache(defaultValues, customDefaults)
    }

    Map getDefaultValues(){
        return defaultValues
    }

    // The reset method is only used for unit tests, in which case the persistence is mocked,
    // that is why no files need to be deleted with this reset.
    static reset(){
        instance = null
    }

    List getCustomDefaults() {
        def result = []
        result.addAll(customDefaults)
        return result
    }

    static void prepare(Script script, Map parameters = [:]) {
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
                if(configFileList.size() > 1) script.echo "Loading configuration file '${configFileName}'"
                def configuration = script.readYaml text: script.libraryResource(configFileName)
                defaultValues = MapUtils.merge(
                        MapUtils.pruneNulls(defaultValues),
                        MapUtils.pruneNulls(configuration))
            }
            DefaultValueCache.createInstance(defaultValues, customDefaults)

            if (script) {
                persistDefaults(script, defaultValues, customDefaults)
            }

        }
    }

    static void persistDefaults(Script script, Map defaultValues, List customDefaults = []) {
        def defaultValuesAbsolutePath = "${script.WORKSPACE}/${defaultValuesRelativePath}"
        def customDefaultsAbsolutePath = "${script.WORKSPACE}/${customDefaultsRelativePath}"

        if (defaultValues && !script.fileExists(defaultValuesAbsolutePath)) {
            def defaultValuesJson = script.readJSON text: groovy.json.JsonOutput.toJson(defaultValues)
            script.writeJSON file: defaultValuesAbsolutePath, json: defaultValuesJson
        }
        if (customDefaults && !script.fileExists(customDefaultsAbsolutePath)) {
            def customDefaultsJson = script.readJSON text: groovy.json.JsonOutput.toJson(customDefaults)
            script.writeJSON file: customDefaultsAbsolutePath, json: customDefaultsJson
        }
    }

    static void createInstanceFromPersistence(Script script) {
        def defaultValues = [:]
        def customDefaults = []
        def defaultValuesAbsolutePath = "${script.WORKSPACE}/${defaultValuesRelativePath}"
        def customDefaultsAbsolutePath = "${script.WORKSPACE}/${customDefaultsRelativePath}"
        if (script.fileExists(defaultValuesAbsolutePath)) {
            defaultValues = script.readJSON file: defaultValuesAbsolutePath, returnPojo: true
        }
        if (script.fileExists(customDefaultsAbsolutePath)) {
            customDefaults = script.readJSON file: customDefaultsAbsolutePath, returnPojo: true
        }
        if (defaultValues) {
            createInstance(defaultValues, customDefaults)
        }
    }
}
