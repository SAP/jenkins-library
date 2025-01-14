package com.sap.piper

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
        if (parameters == null) parameters = [:]
        if (!getInstance() || parameters.customDefaults || parameters.customDefaultsFromFiles) {
            List defaultsFromResources = ['default_pipeline_environment.yml']
            List customDefaults = Utils.appendParameterToStringList(
                [], parameters, 'customDefaults')
            defaultsFromResources.addAll(customDefaults)
            List defaultsFromFiles = Utils.appendParameterToStringList(
                [], parameters, 'customDefaultsFromFiles')

            Map defaultValues = [:]
            defaultValues = addDefaultsFromLibraryResources(steps, defaultValues, defaultsFromResources)
            defaultValues = addDefaultsFromFiles(steps, defaultValues, defaultsFromFiles)

            // The "customDefault" parameter is used for storing which extra defaults need to be
            // passed to piper-go. The library resource 'default_pipeline_environment.yml' shall
            // be excluded, since the go steps have their own in-built defaults in their yaml files.
            createInstance(defaultValues, customDefaults + defaultsFromFiles)
        }
    }

    private static Map addDefaultsFromLibraryResources(Script steps, Map defaultValues, List resourceFiles) {
        for (String configFileName : resourceFiles) {
            if (resourceFiles.size() > 1) {
                steps.echo "Loading configuration file '${configFileName}'"
            }
            Map configuration = steps.readYaml text: steps.libraryResource(configFileName)
            defaultValues = mergeIntoDefaults(defaultValues, configuration)
        }
        return defaultValues
    }

    private static Map addDefaultsFromFiles(Script steps, Map defaultValues, List configFiles) {
        for (String configFileName : configFiles) {
            steps.echo "TESTTEST Loading configuration file '${configFileName}'"
            try {
                Map configuration = steps.readYaml file: ".pipeline/$configFileName"
                defaultValues = mergeIntoDefaults(defaultValues, configuration)
            } catch (Exception e) {
                steps.error "Failed to parse custom defaults as YAML file. " +
                    "Please make sure it is valid YAML, and if loading from a remote location, " +
                    "that the response body only contains valid YAML. " +
                    "If you use a file from a GitHub repository, make sure you've used the 'raw' link, " +
                    "for example https://my.github.local/raw/someorg/shared-config/master/backend-service.yml\n" +
                    "File path: .pipeline/${configFileName}\n" +
                    "Content: ${steps.readFile file: ".pipeline/${configFileName}"}\n" +
                    "Exeption message: ${e.getMessage()}\n" +
                    "Exception stacktrace: ${Arrays.toString(e.getStackTrace())}"
            }
        }
        return defaultValues
    }

    private static Map mergeIntoDefaults(Map defaultValues, Map configuration) {
        return MapUtils.merge(
            MapUtils.pruneNulls(defaultValues),
            MapUtils.pruneNulls(configuration))
    }
}
