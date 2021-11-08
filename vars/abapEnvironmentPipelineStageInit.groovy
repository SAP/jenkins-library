import groovy.transform.Field
import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationLoader

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /**  If set to true, the default scm checkout is skipped */
    'skipCheckout'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
/**
 * This stage initializes the ABAP Environment Pipeline run
 */
void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def stageName = parameters.stageName?:env.STAGE_NAME

    piperStageWrapper (script: script, stageName: stageName, stashContent: [], stageLocking: false, ordinal: 1, telemetryDisabled: true) {

        def skipCheckout = parameters.skipCheckout
        if (skipCheckout != null && !(skipCheckout instanceof Boolean)) {
            error "[${STEP_NAME}] Parameter skipCheckout has to be of type boolean. Instead got '${skipCheckout.class.getName()}'"
        }
        if (!skipCheckout) {
            checkout scm
        }
        setupCommonPipelineEnvironment script: script, customDefaults: parameters.customDefaults

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .addIfEmpty('stageConfigResource', 'com.sap.piper/pipeline/abapEnvironmentPipelineStages.yml')
            .addIfEmpty('stashSettings', 'com.sap.piper/pipeline/abapEnvironmentPipelineStashSettings.yml')
            .withMandatoryProperty('stageConfigResource')
            .use()

        Map stashConfiguration = readYaml(text: libraryResource(config.stashSettings))
        if (config.verbose) echo "Stash config: ${stashConfiguration}"
        script.commonPipelineEnvironment.configuration.stageStashes = stashConfiguration

        //handling of stage and step activation
        piperInitRunStageConfiguration script: script, stageConfigResource: config.stageConfigResource
    }
}
