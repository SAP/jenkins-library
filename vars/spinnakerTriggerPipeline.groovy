import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.JsonUtils
import groovy.text.GStringTemplateEngine

import static com.sap.piper.Prerequisites.checkScript
import groovy.json.JsonOutput
import org.apache.commons.lang3.text.StrSubstitutor


import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils

import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    'spinnaker',
    /**
     * Whether verbose output should be produced.
     * @possibleValues `true`, `false`
     */
    'verbose'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Defines the id of the file credentials in your Jenkins credentials store which contain the client certificate file for Spinnaker authentication.
     * @parentConfigKey spinnaker
     */
    'certFileCredentialsId',
    /**
     * Defines the url of the Spinnaker Gateway Service as API endpoint for communication with Spinnaker.
     * @parentConfigKey spinnaker
     */
    'gateUrl',
    /**
     * Defines the id of the file credentials in your Jenkins credentials store which contain the private key file for Spinnaker authentication.
     * @parentConfigKey spinnaker
     */
    'keyFileCredentialsId',
    /**
     * Defines the name/id of the Spinnaker pipeline.
     * @parentConfigKey spinnaker
     */
    'pipelineNameOrId',
    /**
     * Parameter map containing Spinnaker pipeline parameters.
     * @parentConfigKey spinnaker
     */
    'pipelineParameters',
    /**
     * Defines the timeout in minutes for checking the Spinnaker pipeline result.
     * By setting to `0` the check can be de-activated.
     */
    'timeout'

])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

@Field Map CONFIG_KEY_COMPATIBILITY = [
    application: 'spinnakerApplication',
    certFileCredentialsId: 'certCredentialId',
    gateUrl: 'spinnakerGateUrl',
    keyFileCredentialsId: 'keyCredentialId',
    pipelineNameOrId: 'spinnakerPipeline',
    pipelineParameters: 'pipelineParameters',
    spinnaker: [
        application: 'application',
        certFileCredentialsId: 'certFileCredentialsId',
        keyFileCredentialsId: 'keyFileCredentialsId',
        gateUrl: 'gateUrl',
        pipelineParameters: 'pipelineParameters',
        pipelineNameOrId: 'pipelineNameOrId'
    ]
]

/**
 * Triggers a [Spinnaker](https://spinnaker.io) pipeline from a Jenkins pipeline.
 * Spinnaker is for example used for Continuous Deployment scenarios to various Clouds.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        final script = checkScript(this, parameters) ?: this
        String stageName = parameters.stageName ?: env.STAGE_NAME

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults(CONFIG_KEY_COMPATIBILITY, stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixin(parameters, PARAMETER_KEYS, CONFIG_KEY_COMPATIBILITY)
            .withMandatoryProperty('spinnaker/gateUrl')
            .withMandatoryProperty('spinnaker/application')
            .withMandatoryProperty('spinnaker/pipelineNameOrId')
            .use()

        String paramsString = ""
        if (config.spinnaker.pipelineParameters) {
            def pipelineParameters = [parameters: config.spinnaker.pipelineParameters]

            paramsString = "-d '${new GStringTemplateEngine().createTemplate(JsonOutput.toJson(pipelineParameters)).make([config: config, env: env]).toString()}'"

            if (config.verbose) {
                echo "[${STEP_NAME}] Triggering Spinnaker pipeline with parameters: ${paramsString}"
            }
        }

        def pipelineTriggerResponse

        //ToDO: support userId/pwd authentication or token authentication!

        def curlVerbosity =  (config.verbose) ? '--verbose ' : '--silent '

        withCredentials([
            file(credentialsId: config.spinnaker.keyFileCredentialsId, variable: 'clientKey'),
            file(credentialsId: config.spinnaker.certFileCredentialsId, variable: 'clientCertificate')
        ]) {
            // Trigger a pipeline execution by calling invokePipelineConfigUsingPOST1 (see https://www.spinnaker.io/reference/api/docs.html)
            pipelineTriggerResponse = sh(returnStdout: true, script: "curl -H 'Content-Type: application/json' -X POST ${paramsString} ${curlVerbosity} --cert \$clientCertificate --key \$clientKey ${config.spinnaker.gateUrl}/pipelines/${config.spinnaker.application}/${config.spinnaker.pipelineNameOrId}").trim()
        }
        if (config.verbose) {
            echo "[${STEP_NAME}] Spinnaker pipeline trigger response = ${pipelineTriggerResponse}"
        }

        def pipelineTriggerResponseObj = readJSON text: pipelineTriggerResponse
        if (!pipelineTriggerResponseObj.ref) {
            error "[${STEP_NAME}] Failed to trigger Spinnaker pipeline"
        }

        if (config.timeout == 0) {
            echo "[${STEP_NAME}] Exiting without waiting for Spinnaker pipeline result."
            return
        }

        echo "[${STEP_NAME}] Spinnaker pipeline ${pipelineTriggerResponseObj.ref} triggered, waiting for the pipeline to finish"

        def pipelineStatusResponseObj
        timeout(config.timeout) {
            waitUntil {
                def pipelineStatusResponse
                sleep 10
                withCredentials([
                    file(credentialsId: config.spinnaker.keyFileCredentialsId, variable: 'clientKey'),
                    file(credentialsId: config.spinnaker.certFileCredentialsId, variable: 'clientCertificate')
                ]) {
                    pipelineStatusResponse = sh returnStdout: true, script: "curl -X GET ${config.spinnaker.gateUrl}${pipelineTriggerResponseObj.ref} ${curlVerbosity} --cert \$clientCertificate --key \$clientKey"
                }
                pipelineStatusResponseObj = readJSON text: pipelineStatusResponse
                echo "[${STEP_NAME}] Spinnaker pipeline ${pipelineTriggerResponseObj.ref} status: ${pipelineStatusResponseObj.status}"

                if (pipelineStatusResponseObj.status in ['RUNNING', 'PAUSED', 'NOT_STARTED']) {
                    return false
                } else {
                    return true
                }
            }
        }
        if (pipelineStatusResponseObj.status != 'SUCCEEDED') {
            if (config.verbose) {
                echo "[${STEP_NAME}] Full Spinnaker response = ${new JsonUtils().groovyObjectToPrettyJsonString(pipelineStatusResponseObj)}"
            }
            error "[${STEP_NAME}] Spinnaker pipeline failed with ${pipelineStatusResponseObj.status}"
        }

    }
}
