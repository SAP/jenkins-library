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
            stageToStepMapping.each { stageName, stepsList ->
                containers[stageName] = getContainerForStage(script, stageName as String, stepsList as List, buildTool)
            }
        } catch (Exception e) {
            script.error "Failed to parse container maps in '$yamlResourceName'. It is expected to contain " +
                "a single entry 'containerMaps' in the root, which is a list of stage identifiers. " +
                "Each stage shall contain a list of steps that are known to run in this stage. " +
                "A step entry is either a string (the name of the step), or a list which contains the name " +
                "of the step, and the path to the step metadata resource file (steps implemented in go). " +
                "Error: ${e.getMessage()}"
        }
        script.echo "resulting containers: $containers"
        setMap(containers)
    }

    static Map getContainerForStage(Script script, String stageName, List stepsList, String buildTool) {
        Map containers = [:]
        stepsList.each { stepEntry ->
            String stepName
            String imageName
            if (stepEntry in String) {
                stepName = stepEntry as String
                imageName = getDockerImageNameForGroovyStep(script, stageName, stepName, buildTool)
            } else if (stepEntry in Map) {
                stepName = stepEntry.stepName as String
                String stepMetadata = stepEntry.stepMetadata as String
                imageName = getDockerImageNameForGoStep(script, stageName, stepName, stepMetadata, buildTool)
            } else {
                throw new RuntimeException("Entry '$stepEntry' in containerMaps has unexpected type")
            }
            if (stepName && imageName) {
                containers[imageName] = stepName.toLowerCase()
            }
        }
        return containers
    }

    static String getDockerImageNameForGoStep(Script script, String stageName, String stepName, String stepMetadata, String buildTool) {
        script.echo "Getting docker image name for Go step '$stepName' in stage '$stageName'"
        Map config
        script.withEnv(["STAGE_NAME=$stageName"]) {
            script.piperExecuteBin.prepareExecutionAndGetStepParameters(script, ['buildTool': buildTool], stepMetadata)

            String defaultConfigArgs = script.piperExecuteBin.getCustomDefaultConfigsArg()
            String customConfigArg = script.piperExecuteBin.getCustomConfigArg(script)

            config = script.piperExecuteBin.getStepContextConfig(script, './piper', stepMetadata, defaultConfigArgs, customConfigArg)
        }
        echo "Config for Go step '$stepName': ${config}"
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
