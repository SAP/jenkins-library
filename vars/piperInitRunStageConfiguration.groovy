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
    'verbose'
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
        def currentStage = stage.getKey()
        stage.getValue().stepConditions.each {step ->
            def stepActive = false
            step.getValue().each {condition ->
                Map stepConfig = script.commonPipelineEnvironment.getStepConfiguration(step.getKey(), currentStage)
                switch(condition.getKey()) {
                    case 'config':
                        if (condition.getValue() instanceof Map) {
                            condition.getValue().each {configCondition ->
                                if (MapUtils.getByPath(stepConfig, configCondition.getKey()) in configCondition.getValue()) {
                                    stepActive = true
                                }
                            }
                        } else if (MapUtils.getByPath(stepConfig, condition.getValue())) {
                            stepActive = true
                        }
                        break
                    case 'configKeys':
                        if (condition.getValue() instanceof List) {
                            condition.getValue().each {configKey ->
                                if (MapUtils.getByPath(stepConfig, configKey)) {
                                    stepActive = true
                                }
                            }
                        } else if (MapUtils.getByPath(stepConfig, condition.getValue())) {
                            stepActive = true
                        }
                        break
                    case 'filePatternFromConfig':
                        def conditionValue = MapUtils.getByPath(stepConfig, condition.getValue())
                        if (conditionValue && findFiles(glob: conditionValue)) {
                            stepActive = true
                        }
                        break
                    case 'filePattern':
                        if (findFiles(glob: condition.getValue())) {
                            stepActive = true
                        }
                        break
                    case 'npmScript':
                        def packages = findFiles(glob: '**/package.json', excludes: '**/node_modules/**')
                        for (int i = 0; i < packages.size(); i++) {
                            String packageJsonPath = packages[i].path
                            Map packageJson = readJSON file: packageJsonPath
                            Map npmScripts = packageJson.scripts ?: [:]
                            if (npmScripts[condition.getValue()]) {
                                stepActive = true
                            }
                            break
                        }
                        break
                    case 'extensionExists':
                        stepActive = stepActive || extensionExists(script as Script, config, condition.getValue())
                        break
                }
            }
            script.commonPipelineEnvironment.configuration.runStep."${stage.getKey()}"."${step.getKey()}" = stepActive

            //make sure that also related stage is activated if steps are active
            if (stepActive) script.commonPipelineEnvironment.configuration.runStage[stage.getKey()] = true

        }
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
