package com.sap.piper.k8s

import com.sap.piper.API
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

    void initFromResource(Script script, String yamlResourceName, String buildTool) {
        script.echo "initFromResource(yamlResourceName: $yamlResourceName, buildTool: $buildTool)"
        Map containers = [:]
        boolean[] piperExecutionPrepared = new boolean[1]
        try {
            Map yamlContents = script.readYaml(text: script.libraryResource(yamlResourceName))
            Map stageToStepMapping = yamlContents.containerMaps as Map
            Map stepToMetaDataMapping = yamlContents.stepMetadata as Map ?: [:]
            stageToStepMapping.each { stageName, stepsList ->
                containers[stageName] = getContainersForStage(script, stageName as String, stepsList as List,
                    stepToMetaDataMapping, buildTool, piperExecutionPrepared)
            }
        } catch (Exception e) {
            script.error "Failed to parse container maps in '$yamlResourceName'. It is expected to contain " +
                "the entries 'containerMaps' and optionally 'stepMetadata' in the root." +
                "containerMaps which is a map of stage identifiers to a list of executed steps. " +
                "The optional 'stepMetadata' is a map of (go-implemented) step names to their YAML " +
                "metadata resource file." +
                "Error: ${e.getMessage()}"
        }
        setMap(containers)
    }

    static Map getContainersForStage(Script script, String stageName, List stepsList, Map stepToMetaDataMapping, String buildTool, boolean[] piperExecutionPrepared) {
        Map containers = [:]
        stepsList.each { stepName ->
            String imageName = getDockerImageNameForStepInStage(script, stageName, stepName as String, buildTool)
            String stepMetadata = stepToMetaDataMapping[stepName]
            if (!imageName && stepMetadata) {
                // Retrieve containers for Go steps only if none was found in the config. In this case,
                // a container may still be defined as (conditional) default in the step metadata.
                if (!piperExecutionPrepared[0]) {
                    script.piperExecuteBin.prepareExecution(script)
                    piperExecutionPrepared[0] = true
                }
                imageName = getDockerImageNameForGoStep(script, stageName, stepName as String, stepMetadata, buildTool)
            }
            if (imageName) {
                containers[imageName] = stepName.toLowerCase()
            }
        }
        return containers
    }

    static String getDockerImageNameForStepInStage(Script script, String stageName, String stepName, String buildTool) {
        Map configuration = script.commonPipelineEnvironment.getStepConfiguration(stepName, stageName)

        String dockerImage = configuration.dockerImage

        if (!dockerImage && stepName == "mtaBuild") {
            dockerImage = configuration[configuration.mtaBuildTool]?.dockerImage
        }

        return dockerImage ?: ''
    }

    static String getDockerImageNameForGoStep(Script script, String stageName, String stepName, String stepMetadata, String buildTool) {
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
