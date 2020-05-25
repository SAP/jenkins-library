import groovy.transform.Field
import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationLoader

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = []
@Field Set STEP_CONFIG_KEYS = []
@Field Set PARAMETER_KEYS = []
/**
 * This stage initializes the ABAP Environment Pipeline run
 */
void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    deleteDir()
    checkout scm

    // load default & individual configuration
    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .addIfEmpty('stageConfigResource', 'com.sap.piper/pipeline/abapStageDefaults.yml')
        .use()


    def stageName = parameters.stageName?:env.STAGE_NAME

    piperStageWrapper (script: script, stageName: stageName) {

        setupCommonPipelineEnvironment script: script
        script.commonPipelineEnvironment.configuration.runStage = [:]
        script.commonPipelineEnvironment.configuration.runStep = [:]

        config.stages = (readYaml(text: libraryResource(config.stageConfigResource))).stages

        //handling of stage and step activation
        config.stages.each {stage ->

            //activate stage if stage configuration is available
            if (ConfigurationLoader.stageConfiguration(script, stage.getKey())) {
                script.commonPipelineEnvironment.configuration.runStage[stage.getKey()] = true
            }
        }
    }
}
