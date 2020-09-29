import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils
import groovy.transform.Field
import groovy.text.GStringTemplateEngine

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Allows overriding the Slack Plugin Integration Base Url specified in the global configuration.
     */
    'baseUrl',
    /**
     * Allows overriding of the default massaging channel from the plugin configuration.
     */
    'channel',
    /**
     * Defines the message color`color` defines the message color.
     * @possibleValues one of `good`, `warning`, `danger`, or any hex color code (eg. `#439FE0`)
     */
    'color',
    /**
     * The credentials id for the Slack token.
     * @possibleValues Jenkins credentials id
     */
    'credentialsId',
    /**
     * Send a custom message into the Slack channel.
     */
    'message'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Sends notifications to the Slack channel about the build status.
 *
 * Notification contains:
 *
 * * Build status
 * * Repo Owner
 * * Repo Name
 * * Branch Name
 * * Jenkins Build Number
 * * Jenkins Build URL
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        def utils = parameters.juStabUtils ?: new Utils()
        def script = checkScript(this, parameters) ?: this
        String stageName = parameters.stageName ?: env.STAGE_NAME

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        utils.pushToSWA([step: STEP_NAME], config)

        def buildStatus = script.currentBuild.result
        // resolve templates
        config.color = GStringTemplateEngine.newInstance().createTemplate(config.color).make([buildStatus: buildStatus]).toString()
        if (!config?.message){
            if (!buildStatus) {
                echo "[${STEP_NAME}] currentBuild.result is not set. Skipping Slack notification"
                return
            }
            config.message = GStringTemplateEngine.newInstance().createTemplate(config.defaultMessage).make([buildStatus: buildStatus, env: env]).toString()
        }
        Map options = [:]
        if(config.credentialsId)
            options.put('tokenCredentialId', config.credentialsId)
        for(String entry : ['baseUrl','channel','color','message'])
            if(config.get(entry))
                options.put(entry, config.get(entry))
        slackSend(options)
    }
}
