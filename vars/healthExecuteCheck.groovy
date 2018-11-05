import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import groovy.transform.Field

@Field String STEP_NAME = 'healthExecuteCheck'

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

@Field Set STEP_CONFIG_KEYS = [
    'healthEndpoint',
    'testServerUrl'
]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment,  GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('testServerUrl')
            .use()

        new Utils().pushToSWA([step: STEP_NAME], config)

        def checkUrl = config.testServerUrl
        if(config.healthEndpoint){
            if(!checkUrl.endsWith('/'))
                checkUrl += '/'
            checkUrl += config.healthEndpoint
        }

        def statusCode = curl(checkUrl)
        if (statusCode != '200') {
            error "Health check failed: ${statusCode}"
        } else {
            echo "Health check for ${checkUrl} successful"
        }
    }
}

def curl(url){
    return sh(
        returnStdout: true,
        script: "curl -so /dev/null -w '%{response_code}' ${url}"
    ).trim()
}
