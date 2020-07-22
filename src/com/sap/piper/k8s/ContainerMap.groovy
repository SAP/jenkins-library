package com.sap.piper.k8s

import com.sap.piper.API
import com.sap.piper.Utils
import groovy.json.JsonOutput

@API
@Singleton
class ContainerMap implements Serializable {
    static final long serialVersionUID = 1L

    private Map containerMap = null

    Map getMap() {
        if (containerMap == null) {
            containerMap = [:]
        }
        return containerMap
    }

    void setMap(Map containersMap) {
        containerMap = containersMap
    }

    private static class PiperExecutionPreparator {
        private boolean executionPrepared
        private utils

        public PiperExecutionPreparator(utils){
            this.utils = utils
        }
        void prepareExecution(Script script) {
            if (!executionPrepared) {
                script.piperExecuteBin.prepareExecution(script, utils)
                executionPrepared = true
            }
        }
    }

    void initFromResource(Script script, String yamlResourceName, String buildTool, utils = new Utils()) {
        Map containers = [:]
        PiperExecutionPreparator piperPreparator = new PiperExecutionPreparator(utils)
        Map stageToStepMapping
        Map stepToMetaDataMapping
        try {
            Map yamlContents = script.readYaml(text: script.libraryResource(yamlResourceName))
            stageToStepMapping = yamlContents.containerMaps as Map
            stepToMetaDataMapping = yamlContents.stepMetadata as Map ?: [:]
        } catch (Exception e) {
            script.error "Failed to parse container maps in '$yamlResourceName'. It is expected to contain " +
                "the entries 'containerMaps' and optionally 'stepMetadata' in the root." +
                "containerMaps which is a map of stage identifiers to a list of executed steps. " +
                "The optional 'stepMetadata' is a map of (go-implemented) step names to their YAML " +
                "metadata resource file." +
                "Error: ${e.getMessage()}"
        }
        stageToStepMapping.each { stageName, stepsList ->
            containers[stageName] = getContainersForStage(script, stageName as String, stepsList as List,
                stepToMetaDataMapping, buildTool, piperPreparator)
        }
        setMap(containers)
    }

    static Map getContainersForStage(Script script, String stageName, List stepsList, Map stepToMetaDataMapping, String buildTool, PiperExecutionPreparator piperPreparator) {
        Map containers = [:]
        stepsList.each { stepName ->
            String imageName
            String stepMetadata = stepToMetaDataMapping[stepName]
            if (stepMetadata) {
                piperPreparator.prepareExecution(script)
                imageName = getDockerImageNameForGoStep(script, stageName, stepMetadata, buildTool)
            } else {
                imageName = getDockerImageNameForGroovyStep(script, stageName, stepName as String)
            }
            if (imageName) {
                containers[imageName] = stepName.toLowerCase()
            }
        }
        return containers
    }

    static String getDockerImageNameForGroovyStep(Script script, String stageName, String stepName) {
        Map configuration = script.commonPipelineEnvironment.getStepConfiguration(stepName, stageName)

        String dockerImage = configuration.dockerImage

        if (!dockerImage && stepName == "mtaBuild") {
            dockerImage = configuration[configuration.mtaBuildTool]?.dockerImage
        }

        return dockerImage
    }

    static String getDockerImageNameForGoStep(Script script, String stageName, String stepMetadata, String buildTool) {
        String stepMetadataPath = "metadata/$stepMetadata"
        script.piperExecuteBin.prepareMetadataResource(script, stepMetadataPath)

        Map stepParameters = script.piperExecuteBin.prepareStepParameters(['buildTool': buildTool])

        String defaultConfigArgs = script.piperExecuteBin.getCustomDefaultConfigsArg()
        String customConfigArg = script.piperExecuteBin.getCustomConfigArg(script)

        Map config = [:]
        script.withEnv(["PIPER_parametersJSON=${JsonOutput.toJson(stepParameters)}",
                        "STAGE_NAME=$stageName"]) {
            config = script.piperExecuteBin.getStepContextConfig(script, './piper', stepMetadataPath, defaultConfigArgs,
                customConfigArg)
        }
        return config.dockerImage
    }
}
