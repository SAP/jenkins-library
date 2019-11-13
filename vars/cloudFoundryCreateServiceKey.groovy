import com.sap.piper.GenerateDocumentation
import com.sap.piper.BashUtils
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = 'cloudFoundryCreateServiceKey'

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
         * Cloud Foundry target organization.
         * @parentConfigKey cloudFoundry
         */
        'org',
        /**
         * Cloud Foundry target space.
         * @parentConfigKey cloudFoundry
         */
        'space',
        /**
         * Cloud Foundry service, for which the service key will be created.
         * @parentConfigKey cloudFoundry
         */
        'service',
        /**
         * Cloud Foundry serviceKey, which will be created.
         * @parentConfigKey cloudFoundry
         */
        'serviceKey',
    /** @see dockerExecute */
    'dockerImage',
    /** @see dockerExecute */
    'dockerWorkspace',
    /** @see dockerExecute */
    'stashContent'
]

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Step that creates a service key on cloud foundry
 */
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
            .withMandatoryProperty('cloudFoundry/org')
            .withMandatoryProperty('cloudFoundry/space')
            .withMandatoryProperty('cloudFoundry/credentialsId')
            .withMandatoryProperty('cloudFoundry/service')
            .withMandatoryProperty('cloudFoundry/serviceKey')
            .use()


        executeCreateServiceKey(script, config)
    }
}

private def executeCreateServiceKey(script, Map config) {
    dockerExecute(script:script,dockerImage: config.dockerImage, dockerWorkspace: config.dockerWorkspace) {

        withCredentials([
            usernamePassword(credentialsId: config.cloudFoundry.credentialsId, passwordVariable: 'CF_PASSWORD', usernameVariable: 'CF_USERNAME')
        ]) {
            def returnCode = sh returnStatus: true, script: """#!/bin/bash
            set +x
            set -e
            export HOME=${config.dockerWorkspace}
            cf login -u ${BashUtils.quoteAndEscape(CF_USERNAME)} -p ${BashUtils.quoteAndEscape(CF_PASSWORD)} -a ${config.cloudFoundry.apiEndpoint} -o ${BashUtils.quoteAndEscape(config.cloudFoundry.org)} -s ${BashUtils.quoteAndEscape(config.cloudFoundry.space)};
            cf create-service-key ${BashUtils.quoteAndEscape(config.cloudFoundry.service)} ${BashUtils.quoteAndEscape(config.cloudFoundry.serviceKey)}
            """
            sh "cf logout"
            if (returnCode!=0)  {
                error "[${STEP_NAME}] ERROR: The execution of the create-service-key failed, see the logs above for more details."
                echo "Return Code: $returnCode"
            }
        }
    }
}