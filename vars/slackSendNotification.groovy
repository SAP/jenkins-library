import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import groovy.transform.Field
import groovy.text.SimpleTemplateEngine

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'baseUrl',
    'channel',
    'color',
    'credentialsId',
    'message'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        def utils = parameters.juStabUtils ?: new Utils()
        def script = checkScript(this, parameters) ?: this
        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        new Utils().pushToSWA([step: STEP_NAME], config)

        def buildStatus = script.currentBuild.result
        // resolve templates
        config.color = SimpleTemplateEngine.newInstance().createTemplate(config.color).make([buildStatus: buildStatus]).toString()
        if (!config?.message){
            if (!buildStatus) {
                echo "[${STEP_NAME}] currentBuild.result is not set. Skipping Slack notification"
                return
            }
            config.message = SimpleTemplateEngine.newInstance().createTemplate(config.defaultMessage).make([buildStatus: buildStatus, env: env]).toString()
        }
        Map options = [:]
        if(config.credentialsId)
            options.put('tokenCredentialId', config.credentialsId)
        for(String entry : STEP_CONFIG_KEYS.minus('credentialsId'))
            if(config.get(entry))
                options.put(entry, config.get(entry))
        slackSend(options)
    }
}
