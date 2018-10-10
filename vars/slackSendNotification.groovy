import com.sap.piper.ConfigurationHelper

import groovy.transform.Field
import groovy.text.SimpleTemplateEngine

@Field String STEP_NAME = 'slackSendNotification'
@Field Set STEP_CONFIG_KEYS = ['baseUrl', 'channel', 'color', 'credentialsId', 'message']
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, allowBuildFailure: true) {
        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment, currentBuild: currentBuild]
        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

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
