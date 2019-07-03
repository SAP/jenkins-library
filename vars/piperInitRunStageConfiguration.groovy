import com.sap.piper.DefaultValueCache
import com.sap.piper.ConfigurationLoader

import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.DefaultValueCache
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

    def stageName = parameters.stageName?:env.STAGE_NAME

    DefaultValueCache.getInstance().getProjectConfig().runStage = [:]
    DefaultValueCache.getInstance().getProjectConfig().runStep = [:]

    // load default & individual configuration
    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(GENERAL_CONFIG_KEYS)
        .mixinStepConfig(STEP_CONFIG_KEYS)
        .mixinStageConfig(stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .withMandatoryProperty('stageConfigResource')
        .use()


    config.stages = (readYaml(text: libraryResource(config.stageConfigResource))).stages

    //handling of stage and step activation
    config.stages.each {stage ->

        //activate stage if stage configuration is available
        if (ConfigurationLoader.stageConfiguration(stage.getKey())) {
            DefaultValueCache.getInstance().getProjectConfig().runStage[stage.getKey()] = true
        }
        //-------------------------------------------------------------------------------
        //detailed handling of step and stage activation based on conditions
        DefaultValueCache.getInstance().getProjectConfig().runStep[stage.getKey()] = [:]
        def currentStage = stage.getKey()
        stage.getValue().stepConditions.each {step ->
            def stepActive = false
            step.getValue().each {condition ->
                Map stepConfig = DefaultValueCache.getInstance().commonPipelineEnvironment.getStepConfiguration(step.getKey(), currentStage)
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
                }
            }
            DefaultValueCache.getInstance().getProjectConfig().runStep."${stage.getKey()}"."${step.getKey()}" = stepActive

            //make sure that also related stage is activated if steps are active
            if (stepActive) DefaultValueCache.getInstance().getProjectConfig().runStage[stage.getKey()] = true

        }
    }

    if (config.verbose) {
        echo "[${STEP_NAME}] Debug - Run Stage Configuration: ${DefaultValueCache.getInstance().getProjectConfig().runStage}"
        echo "[${STEP_NAME}] Debug - Run Step Configuration: ${DefaultValueCache.getInstance().getProjectConfig().runStep}"
    }
}
