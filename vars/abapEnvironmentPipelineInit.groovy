import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.GenerateStageDocumentation
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

/**
 * This stage initializes the ABAP Environment Pipeline run
 */
@GenerateStageDocumentation(defaultStageName = 'Init')
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    setupCommonPipelineEnvironment script: script

    script.commonPipelineEnvironment.configuration.runStage = [:]
    script.commonPipelineEnvironment.configuration.runStep = [:]

    // load default & individual configuration
    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        //.withMandatoryProperty('stageConfigResource')
        .addIfEmpty('stageConfigResource', 'com.sap.piper/pipeline/abapStageDefaults.yml')
        .use()


    config.stages = (readYaml(text: libraryResource(config.stageConfigResource))).stages

    //handling of stage and step activation
    config.stages.each {stage ->

        //activate stage if stage configuration is available
        if (ConfigurationLoader.stageConfiguration(script, stage.getKey())) {
            script.commonPipelineEnvironment.configuration.runStage[stage.getKey()] = true
        }
    }
}
