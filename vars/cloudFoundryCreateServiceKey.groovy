import com.sap.piper.GenerateDocumentation
import com.sap.piper.BashUtils
import com.sap.piper.PiperGoUtils
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field def STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/cloudFoundryDeleteService.yaml'

void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters, failOnError: true) {

        def script = checkScript(this, parameters) ?: this

        Map config
        def utils = parameters.juStabUtils ?: new Utils()

        script.commonPipelineEnvironment.writeToDisk(script)

        writeFile(file: METADATA_FILE, text: libraryResource(METADATA_FILE))

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(parameters)}",
        ]) {
            // get context configuration
            config = readJSON (text: sh(returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '${METADATA_FILE}'"))
            // execute step
            dockerExecute(
                script: script,
                dockerImage: config.dockerImage,
                dockerWorkspace: config.dockerWorkspace
            ) {
                withCredentials([usernamePassword(
                    credentialsId: config.cfCredentialsId,
                    passwordVariable: 'PIPER_password',
                    usernameVariable: 'PIPER_username'
                )]) {
                    sh "./piper cloudFoundryCreateServiceKey"
                }
            }
        }
    }
}





/**
 * Step that creates a service key for a specified service instance on Cloud Foundry
 
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this
        Map config = ConfigurationHelper.newInstance(this)
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
*/