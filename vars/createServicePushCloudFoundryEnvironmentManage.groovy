import com.sap.piper.BashUtils
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = 'createServicePushCloudFoundryEnvironmentManage'
@Field Set GENERAL_CONFIG_KEYS = [
    'cfCredentialsId',
    'dockerImage',
    'dockerWorkspace',
    'cloudFoundry', // can contain apiEndpoint,org,space,serviceManifest,manifestVariablesFile, manifestVariablesMap
    'stashContent'
]
@Field Map CONFIG_KEY_COMPATIBILITY = [cloudFoundry: [apiEndpoint: 'cfApiEndpoint', appName:'cfAppName', credentialsId: 'cfCredentialsId', serviceManifest: 'cfServiceManifest', manifestVariablesFiles: 'cfManifestVariablesFiles', manifestVariables: 'cfManifestVariables',  org: 'cfOrg', space: 'cfSpace']]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

def call(Map parameters = [:]) {
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
            .withMandatoryProperty('cloudFoundry/serviceManifest')
            .use()


        utils.pushToSWA([step: STEP_NAME],config)

        utils.unstashAll(config.stashContent)

        if (fileExists(config.cloudFoundry.serviceManifest)) {
            executeCreateServicePush(script, config)
        }
    }
}

private def executeCreateServicePush(script, Map config) {
    dockerExecute(script:script,dockerImage: config.dockerImage, dockerWorkspace: config.dockerWorkspace) {

        String varPart = varOptions(config)

        String varFilePart = varFileOptions(config)

        withCredentials([
            usernamePassword(credentialsId: config.cloudFoundry.credentialsId, passwordVariable: 'CF_PASSWORD', usernameVariable: 'CF_USERNAME')
        ]) {
            sh """#!/bin/bash
            set +x
            set -e
            export HOME=${config.dockerWorkspace}
            cf login -u ${BashUtils.quoteAndEscape(CF_USERNAME)} -p ${BashUtils.quoteAndEscape(CF_PASSWORD)} -a ${config.cloudFoundry.apiEndpoint} -o ${BashUtils.quoteAndEscape(config.cloudFoundry.org)} -s ${BashUtils.quoteAndEscape(config.cloudFoundry.space)};
            cf create-service-push --no-push -f ${BashUtils.quoteAndEscape(config.cloudFoundry.serviceManifest)}${varPart}${varFilePart}
            cf logout
            """
        }
    }
}

private varOptions(Map config) {
    String varPart = ''
    if (config.cloudFoundry.manifestVariables) {
        if (!(config.cloudFoundry.manifestVariables in List)) {
            error "[${STEP_NAME}] ERROR: Parameter config.cloudFoundry.manifestVariables is not a List!"
        }
        config.cloudFoundry.manifestVariables.each {
            if (!(it in Map)) {
                error "[${STEP_NAME}] ERROR: Parameter config.cloudFoundry.manifestVariables.$it is not a Map!"
            }
            it.keySet().each { varKey ->
                String varValue=BashUtils.quoteAndEscape(it.get(varKey).toString())
                varPart += " --var $varKey=$varValue"
            }
        }
    }
    if (varPart) echo "We will add the following string to the cf push call:$varPart !"
    return varPart
}

private String varFileOptions(Map config) {
    String varFilePart = ''
    if (config.cloudFoundry.manifestVariablesFiles) {
        if (!(config.cloudFoundry.manifestVariablesFiles in List)) {
            error "[${STEP_NAME}] ERROR: Parameter config.cloudFoundry.manifestVariablesFiles is not a List!"
        }
        config.cloudFoundry.manifestVariablesFiles.each {
            if (fileExists(it)) {
                varFilePart += " --vars-file ${BashUtils.quoteAndEscape(it)}"
            } else {
                echo "[${STEP_NAME}] [WARNING] We skip adding not-existing file '$it' as a vars-file to the cf create-service-push call"
            }
        }
    }
    if (varFilePart) echo "We will add the following string to the cf push call:$varFilePart !"
    return varFilePart
}

