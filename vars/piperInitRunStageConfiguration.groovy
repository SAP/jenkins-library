import com.sap.piper.ConfigurationLoader

import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.MapUtils
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Print more detailed information into the log.
     * @possibleValues `true`, `false`
     */
    'verbose',
    /**
     * The branch used as productive branch, defaults to master.
     */
    'productiveBranch',
    /**
     * Location for individual stage extensions.
     */
    'projectExtensionsDirectory',
    /**
     * Location for global extensions.
     */
    'globalExtensionsDirectory'

]

@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Defines the library resource that contains the stage configuration settings
     */
    'stageConfigResource'

])

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    String stageName = parameters.stageName ?: env.STAGE_NAME

    script.commonPipelineEnvironment.configuration.runStage = [:]
    script.commonPipelineEnvironment.configuration.runStep = [:]

    // load default & individual configuration
    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults([:], stageName)
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .withMandatoryProperty('stageConfigResource')
        .use()


    config.stages = (readYaml(text: libraryResource(config.stageConfigResource))).stages

    //handling of stage and step activation
    config.stages.each {stage ->

        String currentStage = stage.getKey()
        script.commonPipelineEnvironment.configuration.runStep[currentStage] = [:]

        // Always test step conditions in order to fill runStep[currentStage] map
        boolean anyStepConditionTrue = false
        stage.getValue().stepConditions?.each {step ->
            boolean stepActive = false
            String stepName = step.getKey()
            step.getValue().each {condition ->
                Map stepConfig = script.commonPipelineEnvironment.getStepConfiguration(stepName, currentStage)
                switch(condition.getKey()) {
                    case 'config':
                        stepActive = stepActive || checkConfig(condition, stepConfig)
                        break
                    case 'configKeys':
                        stepActive = stepActive || checkConfigKeys(condition, stepConfig)
                        break
                    case 'filePatternFromConfig':
                        stepActive = stepActive || checkForFilesWithPatternFromConfig(script, condition, stepConfig)
                        break
                    case 'filePattern':
                        stepActive = stepActive || checkForFilesWithPattern(script, condition)
                        break
                    case 'npmScripts':
                        stepActive = stepActive || checkForNpmScriptsInPackages(script, condition)
                        break
                }
            }
            script.commonPipelineEnvironment.configuration.runStep[currentStage][stepName] = stepActive

            anyStepConditionTrue = anyStepConditionTrue || stepActive
        }

        Map stageConfig = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], currentStage)
            .mixinStageConfig(script.commonPipelineEnvironment, currentStage)
            .use()

        boolean runStage
        if (stageConfig.runInAllBranches != true &&
            stage.getValue().onlyProductiveBranch && (config.productiveBranch != env.BRANCH_NAME)) {
            runStage = false
        } else if (ConfigurationLoader.stageConfiguration(script, currentStage)) {
            //activate stage if stage configuration is available
            runStage = true
        } else if (stage.getValue().extensionExists == true) {
            runStage = anyStepConditionTrue || checkExtensionExists(script, config, currentStage)
        } else {
            runStage = anyStepConditionTrue
        }

        script.commonPipelineEnvironment.configuration.runStage[currentStage] = runStage
    }

    if (config.verbose) {
        echo "[${STEP_NAME}] Debug - Run Stage Configuration: ${script.commonPipelineEnvironment.configuration.runStage}"
        echo "[${STEP_NAME}] Debug - Run Step Configuration: ${script.commonPipelineEnvironment.configuration.runStep}"
    }
}

private static boolean checkExtensionExists(Script script, Map config, String stageName) {
    if (!script.piperStageWrapper.allowExtensions(script)) {
        return false
    }
    // NOTE: These keys exist in "config" if they are configured in the general section of the project
    // config or the defaults. However, in piperStageWrapper, these keys could also be configured for
    // the step "piperStageWrapper" to be effective. Don't know if this should be considered here for consistency.
    if (!config.globalExtensionsDirectory && !config.projectExtensionsDirectory) {
        return false
    }
    def projectInterceptorFile = "${config.projectExtensionsDirectory}${stageName}.groovy"
    def globalInterceptorFile = "${config.globalExtensionsDirectory}${stageName}.groovy"
    return script.fileExists(projectInterceptorFile) || script.fileExists(globalInterceptorFile)
}

private static boolean checkConfig(def condition, Map stepConfig) {
    Boolean configExists = false
    if (condition.getValue() instanceof Map) {
        condition.getValue().each {configCondition ->
            if (MapUtils.getByPath(stepConfig, configCondition.getKey()) in configCondition.getValue()) {
                configExists = true
            }
        }
    } else if (MapUtils.getByPath(stepConfig, condition.getValue())) {
        configExists = true
    }
    return configExists
}

private static boolean checkConfigKeys(def condition, Map stepConfig) {
    Boolean configKeyExists = false
    if (condition.getValue() instanceof List) {
        condition.getValue().each { configKey ->
            if (MapUtils.getByPath(stepConfig, configKey)) {
                configKeyExists = true
            }
        }
    } else if (MapUtils.getByPath(stepConfig, condition.getValue())) {
        configKeyExists = true
    }
    return configKeyExists
}

private static boolean checkForFilesWithPatternFromConfig (Script script, def condition, Map stepConfig) {
    def conditionValue = MapUtils.getByPath(stepConfig, condition.getValue())
    if (conditionValue && script.findFiles(glob: conditionValue)) {
        return true
    }
    return false
}

private static boolean checkForFilesWithPattern (Script script, def condition) {
    Boolean filesExist = false
    if (condition.getValue() instanceof List) {
        condition.getValue().each {configKey ->
            if (script.findFiles(glob: configKey)) {
                filesExist = true
            }
        }
    } else {
        if (script.findFiles(glob: condition.getValue())) {
            filesExist = true
        }
    }
    return filesExist
}

private static boolean checkForNpmScriptsInPackages (Script script, def condition) {
    def packages = script.findFiles(glob: '**/package.json', excludes: '**/node_modules/**')
    Boolean npmScriptExists = false
    for (int i = 0; i < packages.size(); i++) {
        String packageJsonPath = packages[i].path
        Map packageJson = script.readJSON file: packageJsonPath
        Map npmScripts = packageJson.scripts ?: [:]
        if (condition.getValue() instanceof List) {
            condition.getValue().each { configKey ->
                if (npmScripts[configKey]) {
                    npmScriptExists = true
                }
            }
        } else {
            if (npmScripts[condition.getValue()]) {
                npmScriptExists = true
            }
        }
    }
    return npmScriptExists
}
