import static com.sap.piper.Prerequisites.checkScript
import static com.sap.piper.BashUtils.quoteAndEscape as q

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

@Field Set STEP_CONFIG_KEYS = [
    /** Optionally with `healthEndpoint` the health function is called if endpoint is not the standard url.*/
    'healthEndpoint',
    /**
     * Health check function is called providing full qualified `testServerUrl` to the health check.
     *
     */
    'testServerUrl'
]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Calls the health endpoint url of the application.
 *
 * The intention of the check is to verify that a suitable health endpoint is available. Such a health endpoint is required for operation purposes.
 *
 * This check is used as a real-life test for your productive health endpoints.
 *
 * !!! note "Check Depth"
 *     Typically, tools performing simple health checks are not too smart. Therefore it is important to choose an endpoint for checking wisely.
 *
 *     This check therefore only checks if the application/service url returns `HTTP 200`.
 *
 *     This is in line with health check capabilities of platforms which are used for example in load balancing scenarios. Here you can find an [example for Amazon AWS](http://docs.aws.amazon.com/elasticloadbalancing/latest/classic/elb-healthchecks.html).
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        String stageName = parameters.stageName ?: env.STAGE_NAME

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment,  GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('testServerUrl')
            .use()

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
        script: "curl -so /dev/null -w '%{response_code}' ${q(url)}"
    ).trim()
}
