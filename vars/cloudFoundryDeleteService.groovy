import com.sap.piper.GenerateDocumentation
import com.sap.piper.BashUtils
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field def STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/cloudFoundryDeleteService.yaml'

@Field Set STEP_CONFIG_KEYS = [
    'cloudFoundry',
        /**
         * Cloud Foundry API endpoint.
         * @parentConfigKey cloudFoundry
         */
        'apiEndpoint',
        /**
         * Credentials to be used for deployment.
         * @parentConfigKey cloudFoundry
         */
        'credentialsId',
        /**
         */
        'org',
        /**
         * Cloud Foundry target space.
         * @parentConfigKey cloudFoundry
         */
        'space',
        /**
        *
        */
        'serviceInstance',
    /** @see dockerExecute */
    'dockerImage',
    /** @see dockerExecute */
    'dockerWorkspace'
]
/* Dominiks Ansatz
@Field Map CONFIG_KEY_COMPATIBILITY = [cloudFoundry: [apiEndpoint: 'cfApiEndpoint', credentialsId: 'cfCredentialsId', org: 'cfOrg', space: 'cfSpace', serviceInstance: 'cfServiceInstance']]
@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()
        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixin(parameters, PARAMETER_KEYS, CONFIG_KEY_COMPATIBILITY)
            .withMandatoryProperty('cloudFoundry/apiEndpoint')
            .withMandatoryProperty('cloudFoundry/credentialsId')
            .withMandatoryProperty('cloudFoundry/org')
            .withMandatoryProperty('cloudFoundry/space')
            .withMandatoryProperty('cloudFoundry/serviceInstance')
            .use()

        deleteService(script, config)
    }
}

private def deleteService(script, Map config) {
    dockerExecute(script:script,dockerImage: config.dockerImage, dockerWorkspace: config.dockerWorkspace) {
        withCredentials([
            usernamePassword(credentialsId: config.cloudFoundry.credentialsId, passwordVariable: 'CF_PASSWORD', usernameVariable: 'CF_USERNAME')
        ]) {
            def returnCode = sh returnStatus: true, script: """#!/bin/bash
            set +x
            set -e
            export HOME=${config.dockerWorkspace}

            ./piper cloudFoundryDeleteService --Username ${BashUtils.quoteAndEscape(CF_USERNAME)} --Password ${BashUtils.quoteAndEscape(CF_PASSWORD)} --API ${BashUtils.quoteAndEscape(config.cloudFoundry.apiEndpoint)} --Space ${BashUtils.quoteAndEscape(config.cloudFoundry.space)} --Organisation ${BashUtils.quoteAndEscape(config.cloudFoundry.org)} --ServiceName ${BashUtils.quoteAndEscape(config.cloudFoundry.serviceInstance)}
            """
            if (returnCode!=0)  {
                error "[${STEP_NAME}] ERROR: The execution of the delete-service plugin failed, see the logs above for more details."
            }
        }
    }
*/


//Daniels Ansatz

//@Field def STEP_NAME = getClass().getName()
//@Field String METADATA_FILE = 'metadata/cloudFoundryDeleteService.yaml'

void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters, failOnError: true) {

        def script = checkScript(this, parameters) ?: this

        Set configKeys = ['dockerImage', 'dockerWorkspace']
        Map jenkinsConfig = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, configKeys)
            .mixinStepConfig(script.commonPipelineEnvironment, configKeys)
            .mixinStageConfig(script.commonPipelineEnvironment, env.STAGE_NAME, configKeys)
            .mixin(parameters, configKeys)
            .use()

        Map config
        def utils = parameters.juStabUtils ?: new Utils()
        parameters.juStabUtils = null

        // telemetry reporting
        utils.pushToSWA([step: STEP_NAME], config)

        //new PiperGoUtils(this, utils).unstashPiperBin()
        //utils.unstash('pipelineConfigAndTests')
        //script.commonPipelineEnvironment.writeToDisk(script)

        writeFile(file: METADATA_FILE, text: libraryResource(METADATA_FILE))

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(parameters)}",
        ]) {
            // get context configuration
            config = readJSON (text: sh(returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '${METADATA_FILE}'"))

            // execute step
            dockerExecute(
                script: script,
                dockerImage: jenkinsConfig.dockerImage,
                dockerWorkspace: jenkinsConfig.dockerWorkspace
            ) {
                /*withCredentials([
                    usernamePassword(credentialsId: config.cloudFoundry.credentialsId, passwordVariable: 'CF_PASSWORD', usernameVariable: 'CF_USERNAME')
                ]) {*/
                withCredentials([usernamePassword(
                    credentialsId: config.credentialsId,
                    passwordVariable: 'PIPER_password',
                    usernameVariable: 'PIPER_username'
                )]) {
                    def returnCode = sh returnStatus: true, script: """#!/bin/bash
                    set +x
                    set -e
                    export HOME=${config.dockerWorkspace}

                    ./piper cloudFoundryDeleteService --Username ${BashUtils.quoteAndEscape(CF_USERNAME)} --Password ${BashUtils.quoteAndEscape(CF_PASSWORD)} --API ${BashUtils.quoteAndEscape(config.cloudFoundry.apiEndpoint)} --Space ${BashUtils.quoteAndEscape(config.cloudFoundry.space)} --Organisation ${BashUtils.quoteAndEscape(config.cloudFoundry.org)} --ServiceName ${BashUtils.quoteAndEscape(config.cloudFoundry.serviceInstance)}
                    """
                    if (returnCode!=0)  {
                        error "[${STEP_NAME}] ERROR: The execution of the delete-service plugin failed, see the logs above for more details."
                    }
                }
            }
        }
    }
}



