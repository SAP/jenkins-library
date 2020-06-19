package com.sap.piper.k8s

import com.sap.piper.API
import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger
import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils

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
        try {
            Map yamlContents = script.readYaml(text: script.libraryResource(yamlResourceName))
            Map stageToStepMapping = yamlContents.containerMaps as Map
            Map stepToMetaDataMapping = yamlContents.stepMetadata as Map ?: [:]
            stageToStepMapping.each { stageName, stepsList ->
                containers[stageName] = getContainerForStage(script, stageName as String, stepsList as List, stepToMetaDataMapping, buildTool)
            }
        } catch (Exception e) {
            script.error "Failed to parse container maps in '$yamlResourceName'. It is expected to contain " +
                "the entries 'containerMaps' and optionally 'stepMetadata' in the root." +
                "containerMaps which is a map of stage identifiers to a list of executed steps. " +
                "The optional 'stepMetadata' is a map of (go-implemented) step names to their YAML " +
                "metadata resource file." +
                "Error: ${e.getMessage()}"
        }
        script.echo "resulting containers: $containers"
        setMap(containers)
    }

    static Map getContainerForStage(Script script, String stageName, List stepsList, Map stepToMetaDataMapping, String buildTool) {
        Map containers = [:]
        stepsList.each { stepName ->
            String imageName
            String stepMetadata = stepToMetaDataMapping[stepName]
            if (stepMetadata) {
                imageName = getDockerImageNameForGoStep(script, stageName, stepName as String, stepMetadata, buildTool)
            } else {
                imageName = getDockerImageNameForGroovyStep(script, stageName, stepName as String, buildTool)
            }
            if (imageName) {
                containers[imageName] = stepName.toLowerCase()
            }
        }
        return containers
    }

    static String getDockerImageNameForGoStep(Script script, String stageName, String stepName, String stepMetadata, String buildTool) {
        script.echo "Getting docker image name for Go step '$stepName' in stage '$stageName'"

        String stepMetadataPath = "metadata/$stepMetadata"

        Map stepParameters = script.piperExecuteBin.prepareExecutionAndGetStepParameters(script, ['buildTool': buildTool], stepMetadataPath)

        String defaultConfigArgs = script.piperExecuteBin.getCustomDefaultConfigsArg()
        String customConfigArg = script.piperExecuteBin.getCustomConfigArg(script)

        Map config = [:]
        script.withEnv(["PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(stepParameters)}",
                        "STAGE_NAME=$stageName"]) {
            config = script.piperExecuteBin.getStepContextConfig(script, './piper', stepMetadataPath, defaultConfigArgs, customConfigArg)
        }
        script.echo "Config for Go step '$stepName': ${config}"
        return config.dockerImage
    }

    static String getDockerImageNameForGroovyStep(Script script, String stageName, String stepName, String buildTool) {
        script.echo "Getting docker image name for Groovy step '$stepName' in stage '$stageName'"
        Map configuration = loadEffectiveStepConfigurationInStage(script, stageName, stepName)
        String dockerImage = configuration.dockerImage

        if(!dockerImage && stepName == "mtaBuild"){
            dockerImage = configuration[configuration.mtaBuildTool]?.dockerImage
        }

        return dockerImage ?: ''
    }

    private static Map loadEffectiveStepConfigurationInStage(Script script, String stageName, String stepName) {
        final Map stageConfiguration = loadEffectiveStageConfiguration(script, stageName)
        final Map stepConfiguration = loadEffectiveStepConfiguration(script, stepName)
        return ConfigurationMerger.merge(stageConfiguration, null, stepConfiguration)
    }

    private static Map loadEffectiveStepConfiguration(Script script, String stepName) {
        Map stepConfiguration = ConfigurationLoader.stepConfiguration(script, stepName)
        Map defaultStepConfiguration = ConfigurationLoader.defaultStepConfiguration(script, stepName)
        return ConfigurationMerger.merge(stepConfiguration, null, defaultStepConfiguration)
    }

    private static Map loadEffectiveStageConfiguration(Script script, String stageName) {
        Map stageConfiguration = ConfigurationLoader.stageConfiguration(script, stageName)
        Map defaultStageConfiguration = ConfigurationLoader.defaultStageConfiguration(script, stageName)
        return ConfigurationMerger.merge(stageConfiguration, null, defaultStageConfiguration)
    }
}
