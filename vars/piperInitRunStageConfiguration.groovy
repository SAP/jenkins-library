import com.sap.piper.ConfigurationLoader

import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
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
                switch(condition.getKey()) {
                    case 'config':
                        if (condition.getValue() instanceof Map) {
                            condition.getValue().each {configCondition ->
                                if (script.commonPipelineEnvironment.getStepConfiguration(step.getKey(), currentStage)?.get(configCondition.getKey()) in configCondition.getValue()) {
                                    stepActive = true
                                }
                            }
                        } else if (script.commonPipelineEnvironment.getStepConfiguration(step.getKey(), currentStage)?.get(condition.getValue())) {
                            stepActive = true
                        }
                        break
                    case 'filePatternFromConfig':
                        def conditionValue=script.commonPipelineEnvironment.getStepConfiguration(step.getKey(), currentStage)?.get(condition.getValue())
                        if (conditionValue && findFiles(glob: conditionValue)) {
                            stepActive = true
                        }
                        break
                    case 'filePattern':
                        if (findFiles(glob: condition.getValue())) {
                            stepActive = true
                        }
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
