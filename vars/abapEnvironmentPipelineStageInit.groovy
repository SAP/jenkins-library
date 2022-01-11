import com.sap.piper.Utils
import groovy.transform.Field
import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationLoader

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /**  If set to true, the default scm checkout is skipped */
    'skipCheckout',
    /**
     * Optional list of file paths or URLs, which must point to YAML content. These work exactly like
     * `customDefaults`, but from local or remote files instead of library resources. They are merged with and
     * take precedence over `customDefaults`.
     */
    'customDefaultsFromFiles',
    /**
     * Mandatory if you skip the checkout. Then you need to unstash your workspace to get the e.g. configuration.
     */
    'stashContent'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
/**
 * This stage initializes the ABAP Environment Pipeline run
 */
void call(Map parameters = [:]) {
    def utils = parameters.juStabUtils ?: new Utils()
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
        else {
            def stashContent = parameters.stashContent
            if(stashContent == null || stashContent.size() == 0) {
                error "[${STEP_NAME}] needs stashes if you skip checkout"
            }
            utils.unstashAll(stashContent)
        }
        setupCommonPipelineEnvironment script: script,
            customDefaults: parameters.customDefaults,
            customDefaultsFromFiles: parameters.customDefaultsFromFiles

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
