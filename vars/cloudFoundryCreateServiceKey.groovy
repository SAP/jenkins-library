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
         * Cloud Foundry credentials.
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
         * Cloud Foundry service instance, for which the service key will be created.
         * @parentConfigKey cloudFoundry
         */
        'serviceInstance',
        /**
         * Cloud Foundry service key, which will be created.
         * @parentConfigKey cloudFoundry
         */
        'serviceKey',
        /**
         * Cloud Foundry service key configuration.
         * @parentConfigKey cloudFoundry
         */
        'serviceKeyConfig',
    /** @see dockerExecute */
    'dockerImage',
    /** @see dockerExecute */
    'dockerWorkspace'
]

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
@Field Map CONFIG_KEY_COMPATIBILITY = [cloudFoundry: [apiEndpoint: 'cfApiEndpoint', credentialsId: 'cfCredentialsId', org: 'cfOrg', space: 'cfSpace', serviceInstance: 'cfServiceInstance', serviceKey: 'cfServiceKey', serviceKeyConfig: 'cfServiceKeyConfig']]

/**
 * Step that creates a service key for a specified service instance on Cloud Foundry
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this
        Map config = ConfigurationHelper.newInstance(this, script)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixin(parameters, PARAMETER_KEYS, CONFIG_KEY_COMPATIBILITY)
            .withMandatoryProperty('cloudFoundry/org')
            .withMandatoryProperty('cloudFoundry/space')
            .withMandatoryProperty('cloudFoundry/credentialsId')
            .withMandatoryProperty('cloudFoundry/serviceInstance')
            .withMandatoryProperty('cloudFoundry/serviceKey')
            .withMandatoryProperty('cloudFoundry/apiEndpoint')
            .use()

        echo "[${STEP_NAME}] Info: docker image: ${config.dockerImage}, docker workspace: ${config.dockerWorkspace}"
        executeCreateServiceKey(script, config)
    }
}

private def executeCreateServiceKey(script, Map config) {
    dockerExecute(script:script, dockerImage: config.dockerImage, dockerWorkspace: config.dockerWorkspace) {

        withCredentials([
            usernamePassword(credentialsId: config.cloudFoundry.credentialsId, passwordVariable: 'CF_PASSWORD', usernameVariable: 'CF_USERNAME')
        ]) {
            String flag = config.cloudFoundry.serviceKeyConfig == null ? "" : "-c"
            String serviceKeyConfig = config.cloudFoundry.serviceKeyConfig == null ? "" : config.cloudFoundry.serviceKeyConfig
            bashScript =
                """#!/bin/bash
                set +x
                set -e
                export HOME=${config.dockerWorkspace}
                cf login -u ${BashUtils.quoteAndEscape(CF_USERNAME)} -p ${BashUtils.quoteAndEscape(CF_PASSWORD)} -a ${config.cloudFoundry.apiEndpoint} -o ${BashUtils.quoteAndEscape(config.cloudFoundry.org)} -s ${BashUtils.quoteAndEscape(config.cloudFoundry.space)};
                cf create-service-key ${BashUtils.quoteAndEscape(config.cloudFoundry.serviceInstance)} ${BashUtils.quoteAndEscape(config.cloudFoundry.serviceKey)} ${flag} ${BashUtils.quoteAndEscape(serviceKeyConfig)}
                """
            def returnCode = sh returnStatus: true, script: bashScript
            sh "cf logout"
            if (returnCode!=0)  {
                error "[${STEP_NAME}] Error: The execution of create-service-key failed, see the logs above for more details."
                echo "Return Code: $returnCode"
            }
        }
    }
}
