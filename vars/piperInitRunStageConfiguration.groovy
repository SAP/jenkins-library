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
    def stageName = parameters.stageName?:env.STAGE_NAME

    script.commonPipelineEnvironment.configuration.runStage = [:]
    script.commonPipelineEnvironment.configuration.runStep = [:]

    // load default & individual configuration
    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .withMandatoryProperty('stageConfigResource')
        .use()


    config.stages = (readYaml(text: libraryResource(config.stageConfigResource))).stages

    //handling of stage and step activation
    config.stages.each {stage ->

        //activate stage if stage configuration is available
        if (ConfigurationLoader.stageConfiguration(script, stage.getKey())) {
            script.commonPipelineEnvironment.configuration.runStage[stage.getKey()] = true
        }
        //-------------------------------------------------------------------------------
        //detailed handling of step and stage activation based on conditions
        script.commonPipelineEnvironment.configuration.runStep[stage.getKey()] = [:]
        String currentStage = stage.getKey()
        boolean anyStepConditionTrue = false
        stage.getValue().stepConditions.each {step ->
            def stepActive = false
            step.getValue().each {condition ->
                Map stepConfig = script.commonPipelineEnvironment.getStepConfiguration(step.getKey(), currentStage)
                switch(condition.getKey()) {
                    case 'config':
                        stepActive |= checkConfig(condition, stepConfig)
                        break
                    case 'configKeys':
                        stepActive |= checkConfigKeys(condition, stepConfig)
                        break
                    case 'filePatternFromConfig':
                        stepActive |= checkForFilesWithPatternFromConfig(script, condition, stepConfig)
                        break
                    case 'filePattern':
                        stepActive |= checkForFilesWithPattern(script, condition)
                        break
                    case 'npmScripts':
                        stepActive |= checkForNpmScriptsInPackages(script, condition)
                        break
                }
            }
            script.commonPipelineEnvironment.configuration.runStep."${currentStage}"."${step.getKey()}" = stepActive

            anyStepConditionTrue |= stepActive

        }
        boolean runStage = anyStepConditionTrue
        if (stage.getValue().extensionExists) {
            runStage |= extensionExists(script as Script, config, currentStage)
        }

        if (stage.getValue().onlyProductiveBranch && (config.productiveBranch != env.BRANCH_NAME)) {
            runStage = false
        }

        script.commonPipelineEnvironment.configuration.runStage[currentStage] = runStage
    }

    if (config.verbose) {
        echo "[${STEP_NAME}] Debug - Run Stage Configuration: ${script.commonPipelineEnvironment.configuration.runStage}"
        echo "[${STEP_NAME}] Debug - Run Step Configuration: ${script.commonPipelineEnvironment.configuration.runStep}"
    }
}

private static boolean extensionExists(Script script, Map config, def stageName) {
    if (!stageName || !(stageName in CharSequence)) {
        return false
    }
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
