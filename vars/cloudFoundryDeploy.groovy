import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper

import groovy.transform.Field

@Field String STEP_NAME = 'cloudFoundryDeploy'
@Field Set STEP_CONFIG_KEYS_COMPATIBILITY = [
    'cfApiEndpoint',
    'cfAppName',
    'cfCredentialsId',
    'cfManifest',
    'cfOrg',
    'cfSpace'
]
@Field Set STEP_CONFIG_KEYS = STEP_CONFIG_KEYS_COMPATIBILITY + [
    'cloudFoundry',
    'deployUser',
    'deployTool',
    'deployType',
    'dockerImage',
    'dockerWorkspace',
    'mtaDeployParameters',
    'mtaExtensionDescriptor',
    'mtaPath',
    'smokeTestScript',
    'smokeTestStatusCode',
    'stashContent']
@Field Map CONFIG_KEY_COMPATIBILITY = [cloudFoundry: [apiEndpoint: 'cfApiEndpoint', appName:'cfAppName', credentialsId: 'cfCredentialsId', manifest: 'cfManifest', org: 'cfOrg', space: 'cfSpace']]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

def call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def utils = parameters.juStabUtils
        if (utils == null) {
            utils = new Utils()
        }

        def script = parameters.script
        if (script == null)
            script = [commonPipelineEnvironment: commonPipelineEnvironment]

        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .dependingOn('deployTool').mixin('dockerImage')
            .dependingOn('deployTool').mixin('dockerWorkspace')
            .handleCompatibility(this, CONFIG_KEY_COMPATIBILITY)
            //.withMandatoryProperty('cloudFoundry.Org')
            //.withMandatoryProperty('cloudFoundry.Space')
            .use()

        echo "[${STEP_NAME}] General parameters: deployTool=${config.deployTool}, deployType=${config.deployType}, cfApiEndpoint=${config.cloudFoundry.apiEndpoint}, cfOrg=${config.cloudFoundry.org}, cfSpace=${config.cloudFoundry.space}, cfCredentialsId=${config.cloudFoundry.credentialsId}, deployUser=${config.deployUser}"

        utils.unstash 'deployDescriptor'

        if (config.deployTool == 'mtaDeployPlugin') {
            // set default mtar path
            config = new ConfigurationHelper(config)
                .addIfEmpty('mtaPath', config.mtaPath?:findMtar())
                .use()

            dockerExecute(dockerImage: config.dockerImage, dockerWorkspace: config.dockerWorkspace, stashContent: config.stashContent) {
                deployMta(config)
            }
            return
        }

        if (config.deployTool == 'cf_native') {
            def smokeTest = ''

            if (config.smokeTestScript == 'blueGreenCheck.sh') {
                writeFile file: config.smokeTestScript, text: libraryResource(config.smokeTestScript)
            } else {
                utils.unstash 'pipelineConfigAndTests'
            }
            smokeTest = '--smoke-test $(pwd)/' + config.smokeTestScript
            sh "chmod +x ${config.smokeTestScript}"

            echo "[${STEP_NAME}] CF native deployment (${config.deployType}) with cfAppName=${config.cloudFoundry.appName}, cfManifest=${config.cloudFoundry.manifest}, smokeTestScript=${config.smokeTestScript}"

            dockerExecute (dockerImage: config.dockerImage, dockerWorkspace: config.dockerWorkspace, stashContent: config.stashContent, dockerEnvVars: [CF_HOME:"${config.dockerWorkspace}", CF_PLUGIN_HOME:"${config.dockerWorkspace}", STATUS_CODE: "${config.smokeTestStatusCode}"]) {
                deployCfNative(config)
            }

            return
        }
    }
}

def findMtar(){
    def mtarPath = ''
    def mtarFiles = findFiles(glob: '**/target/*.mtar')

    if(mtarFiles.length > 1){
        error 'Found multiple *.mtar files, please specify file via mtaPath parameter! ${mtarFiles}'
    }
    if(mtarFiles.length == 1){
        return mtarFiles[0].path
    }
    error 'No *.mtar file found!'
}

def deployCfNative (config) {
    withCredentials([usernamePassword(
        credentialsId: config.cloudFoundry.credentialsId,
        passwordVariable: 'password',
        usernameVariable: 'username'
    )]) {
        def deployCommand = 'push'
        if (config.deployType == 'blue-green') {
            deployCommand = 'blue-green-deploy'
        } else {
            config.smokeTest = ''
        }

        // check if appName is available
        if (config.cloudFoundry.appName == null || config.cloudFoundry.appName == '') {
            if (fileExists(config.cloudFoundry.manifest)) {
                def manifest = readYaml file: config.cloudFoundry.manifest
                if (!manifest || !manifest.applications || !manifest.applications[0].name)
                    error "[${STEP_NAME}] ERROR: No appName available in manifest ${config.cloudFoundry.manifest}."

            } else {
                error "[${STEP_NAME}] ERROR: No manifest file ${config.cloudFoundry.manifest} found."
            }
        }

        sh """#!/bin/bash
            set +x
            export HOME=/home/piper
            cf login -u \"${username}\" -p '${password}' -a ${config.cloudFoundry.apiEndpoint} -o \"${config.cloudFoundry.org}\" -s \"${config.cloudFoundry.space}\"
            cf plugins
            cf ${deployCommand} ${config.cloudFoundry.appName?"\"${config.cloudFoundry.appName}\"":''} -f \"${config.cloudFoundry.manifest}\" ${config.smokeTest}"""
        def retVal = sh script: "cf app \"${config.cloudFoundry.appName}-old\"", returnStatus: true
        if (retVal == 0) {
            sh "cf delete \"${config.cloudFoundry.appName}-old\" -r -f"
        }
        sh "cf logout"
    }
}

def deployMta (config) {
    if (config.mtaExtensionDescriptor == null) config.mtaExtensionDescriptor = ''
    if (!config.mtaExtensionDescriptor.isEmpty() && !config.mtaExtensionDescriptor.startsWith('-e ')) config.mtaExtensionDescriptor = "-e ${config.mtaExtensionDescriptor}"

    def deployCommand = 'deploy'
    if (config.deployType == 'blue-green')
        deployCommand = 'bg-deploy'

    withCredentials([usernamePassword(
        credentialsId: config.cloudFoundry.credentialsId,
        passwordVariable: 'password',
        usernameVariable: 'username'
    )]) {
        echo "[${STEP_NAME}] Deploying MTA (${config.mtaPath}) with following parameters: ${config.mtaExtensionDescriptor} ${config.mtaDeployParameters}"
        sh """#!/bin/bash
            export HOME=/home/piper
            set +x
            cf api ${config.cloudFoundry.apiEndpoint}
            cf login -u ${username} -p '${password}' -a ${config.cloudFoundry.apiEndpoint} -o \"${config.cloudFoundry.org}\" -s \"${config.cloudFoundry.space}\"
            cf plugins
            cf ${deployCommand} ${config.mtaPath} ${config.mtaDeployParameters} ${config.mtaExtensionDescriptor}"""
        sh "cf logout"
    }
}
