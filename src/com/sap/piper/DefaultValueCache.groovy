package com.sap.piper

import com.sap.piper.MapUtils

import com.cloudbees.groovy.cps.NonCPS

@API
class DefaultValueCache implements Serializable {
    private static DefaultValueCache instance

    private static String DEFAULT_PROJECT_CONFIG_FILE_PATH = '.pipeline/config.yml'

    private Map defaultValues
    private Map projectConfig

    private DefaultValueCache(Map defaultValues, Map projectConfig){
        this.defaultValues = defaultValues
        this.projectConfig = projectConfig
    }

    @NonCPS
    static getInstance(){
        return instance
    }

    static createInstance(Map defaultValues, Map projectConfig){
        instance = new DefaultValueCache(defaultValues, projectConfig)
    }

    @NonCPS
    Map getDefaultValues(){
        return defaultValues
    }

    @NonCPS
    Map getProjectConfig() {
        return projectConfig
    }

    static reset(){
        instance = null
    }

    static void prepare(Script steps, Map parameters = [:]) {
        if(parameters == null) parameters = [:]
        if(!DefaultValueCache.getInstance() || parameters.customDefaults) {
            def defaultValues = [:]
            def configFileList = ['default_pipeline_environment.yml']
            def customDefaults = parameters.customDefaults

            if(customDefaults in String)
                customDefaults = [customDefaults]
            if(customDefaults in List)
                configFileList += customDefaults
            for (def configFileName : configFileList){
                if(configFileList.size() > 1) steps.echo "Loading configuration file '${configFileName}'"
                def configuration = steps.readYaml text: steps.libraryResource(configFileName)
                defaultValues = MapUtils.merge(
                        MapUtils.pruneNulls(defaultValues),
                        MapUtils.pruneNulls(configuration))
            }

            def projectConfigFileName = parameters.projectConfig ?: DEFAULT_PROJECT_CONFIG_FILE_PATH
            boolean projectConfigFileExists = steps.fileExists projectConfigFileName

            def projectConfig

            if(projectConfigFileExists) {
                projectConfig = steps.readYaml file: projectConfigFileName
            } else {
                if(projectConfigFileName != DEFAULT_PROJECT_CONFIG_FILE_PATH)
                    steps.error("Explicitly configured project config file '${projectConfigFileName}' does not exist.")
                projectConfig = [:]
            }

            DefaultValueCache.createInstance(defaultValues, projectConfig)
        }
    }
}
