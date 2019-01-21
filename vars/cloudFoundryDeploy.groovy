import com.sap.piper.JenkinsUtils

import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
import com.sap.piper.CfManifestUtils

import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

@Field Set STEP_CONFIG_KEYS = [
    'cloudFoundry',
    'deployUser',
    'deployTool',
    'deployType',
    'keepOldInstance',
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

void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def utils = parameters.juStabUtils ?: new Utils()
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()

        def script = checkScript(this, parameters)
        if (script == null)
            script = this

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixin(parameters, PARAMETER_KEYS, CONFIG_KEY_COMPATIBILITY)
            .dependingOn('deployTool').mixin('dockerImage')
            .dependingOn('deployTool').mixin('dockerWorkspace')
            .withMandatoryProperty('cloudFoundry/org')
            .withMandatoryProperty('cloudFoundry/space')
            .withMandatoryProperty('cloudFoundry/credentialsId')
            .use()

        utils.pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'deployTool',
            stepParam1: config.deployTool,
            stepParamKey2: 'deployType',
            stepParam2: config.deployType,
            stepParamKey3: 'scriptMissing',
            stepParam3: parameters?.script == null
        ], config)

        echo "[${STEP_NAME}] General parameters: deployTool=${config.deployTool}, deployType=${config.deployType}, cfApiEndpoint=${config.cloudFoundry.apiEndpoint}, cfOrg=${config.cloudFoundry.org}, cfSpace=${config.cloudFoundry.space}, cfCredentialsId=${config.cloudFoundry.credentialsId}, deployUser=${config.deployUser}"

        //make sure that all relevant descriptors, are available in workspace
        utils.unstashAll(config.stashContent)
        //make sure that for further execution whole workspace, e.g. also downloaded artifacts are considered
        config.stashContent = []

        boolean deploy = false
        boolean deploySuccess = true
        try {
            if (config.deployTool == 'mtaDeployPlugin') {
                deploy = true
                // set default mtar path
                config = ConfigurationHelper.newInstance(this, config)
                    .addIfEmpty('mtaPath', config.mtaPath?:findMtar())
                    .use()

                dockerExecute(script: script, dockerImage: config.dockerImage, dockerWorkspace: config.dockerWorkspace, stashContent: config.stashContent) {
                    deployMta(config)
                }
            }

            if (config.deployTool == 'cf_native') {
                deploy = true
                config.smokeTest = ''

                if (config.smokeTestScript == 'blueGreenCheckScript.sh') {
                    writeFile file: config.smokeTestScript, text: libraryResource(config.smokeTestScript)
                }

                config.smokeTest = '--smoke-test $(pwd)/' + config.smokeTestScript
                sh "chmod +x ${config.smokeTestScript}"

                echo "[${STEP_NAME}] CF native deployment (${config.deployType}) with cfAppName=${config.cloudFoundry.appName}, cfManifest=${config.cloudFoundry.manifest}, smokeTestScript=${config.smokeTestScript}"

                dockerExecute (
                    script: script,
                    dockerImage: config.dockerImage,
                    dockerWorkspace: config.dockerWorkspace,
                    stashContent: config.stashContent,
                    dockerEnvVars: [CF_HOME:"${config.dockerWorkspace}", CF_PLUGIN_HOME:"${config.dockerWorkspace}", STATUS_CODE: "${config.smokeTestStatusCode}"]
                ) {
                    deployCfNative(config)
                }
            }
        } catch (err) {
            deploySuccess = false
            throw err
        } finally {
            if (deploy) {
                reportToInflux(script, config, deploySuccess, jenkinsUtils)
            }

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
        def deployCommand = selectCfDeployCommandForDeployType(config)

        if (config.deployType == 'blue-green') {
            handleLegacyCfManifest(config)
        } else {
            config.smokeTest = ''
        }

        def blueGreenDeployOptions = deleteOptionIfRequired(config)

        // check if appName is available
        if (config.cloudFoundry.appName == null || config.cloudFoundry.appName == '') {
            if (config.deployType == 'blue-green') {
                error "[${STEP_NAME}] ERROR: Blue-green plugin requires app name to be passed (see https://github.com/bluemixgaragelondon/cf-blue-green-deploy/issues/27)"
            }
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
            set -e
            export HOME=${config.dockerWorkspace}
            cf login -u \"${username}\" -p '${password}' -a ${config.cloudFoundry.apiEndpoint} -o \"${config.cloudFoundry.org}\" -s \"${config.cloudFoundry.space}\"
            cf plugins
            cf ${deployCommand} ${config.cloudFoundry.appName ?: ''} ${blueGreenDeployOptions} -f '${config.cloudFoundry.manifest}' ${config.smokeTest}
            ${stopOldAppIfRequired(config)}
            """
        sh "cf logout"
    }
}

private String selectCfDeployCommandForDeployType(Map config) {
    if (config.deployType == 'blue-green') {
        return 'blue-green-deploy'
    } else {
        return 'push'
    }
}

private String deleteOptionIfRequired(Map config) {
    boolean deleteOldInstance = !config.keepOldInstance
    if (deleteOldInstance && config.deployType == 'blue-green') {
        return '--delete-old-apps'
    } else {
        return ''
    }
}

private String stopOldAppIfRequired(Map config) {
    if (config.keepOldInstance && config.deployType == 'blue-green') {
        return "cf stop ${config.cloudFoundry.appName}-old"
    } else {
        return ''
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
            export HOME=${config.dockerWorkspace}
            set +x
            set -e
            cf api ${config.cloudFoundry.apiEndpoint}
            cf login -u ${username} -p '${password}' -a ${config.cloudFoundry.apiEndpoint} -o \"${config.cloudFoundry.org}\" -s \"${config.cloudFoundry.space}\"
            cf plugins
            cf ${deployCommand} ${config.mtaPath} ${config.mtaDeployParameters} ${config.mtaExtensionDescriptor}"""
        sh "cf logout"
    }
}

def handleLegacyCfManifest(config) {
    def manifest = readYaml file: config.cloudFoundry.manifest
    String originalManifest = manifest.toString()
    manifest = CfManifestUtils.transform(manifest)
    String transformedManifest = manifest.toString()
    if (originalManifest != transformedManifest) {
        echo """The file ${config.cloudFoundry.manifest} is not compatible with the Cloud Foundry blue-green deployment plugin. Re-writing inline.
See this issue if you are interested in the background: https://github.com/cloudfoundry/cli/issues/1445.\n
Original manifest file content: $originalManifest\n
Transformed manifest file content: $transformedManifest"""
        sh "rm ${config.cloudFoundry.manifest}"
        writeYaml file: config.cloudFoundry.manifest, data: manifest
    }
}


private void reportToInflux(script, config, deploySuccess, JenkinsUtils jenkinsUtils) {
    def deployUser = ''
    withCredentials([usernamePassword(
        credentialsId: config.cloudFoundry.credentialsId,
        passwordVariable: 'password',
        usernameVariable: 'username'
    )]) {
        deployUser = username
    }

    def timeFinished = new Date().format( 'MMM dd, yyyy - HH:mm:ss' )
    def triggerCause = jenkinsUtils.isJobStartedByUser()?'USER':(jenkinsUtils.isJobStartedByTimer()?'TIMER': 'OTHER')

    def deploymentData = [deployment_data: [
        artifactUrl: 'n/a', //might be added later on during pipeline run (written to commonPipelineEnvironment)
        deployTime: timeFinished,
        jobTrigger: triggerCause
    ]]
    def deploymentDataTags = [deployment_data: [
        artifactVersion: script.commonPipelineEnvironment.getArtifactVersion(),
        deployUser: deployUser,
        deployResult: deploySuccess?'SUCCESS':'FAILURE',
        cfApiEndpoint: config.cloudFoundry.apiEndpoint,
        cfOrg: config.cloudFoundry.org,
        cfSpace: config.cloudFoundry.space,
    ]]
    influxWriteData script: script, customData: [:], customDataTags: [:], customDataMap: deploymentData, customDataMapTags: deploymentDataTags
}
