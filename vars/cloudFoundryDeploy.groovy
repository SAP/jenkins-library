import com.sap.piper.JenkinsUtils

import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
import com.sap.piper.CfManifestUtils

import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

@Field Set STEP_CONFIG_KEYS = [
    'cloudFoundry',
        /**
         * Cloud Foundry API endpoint.
         * @parentConfigKey cloudFoundry
         */
        'apiEndpoint',
        /**
         * Defines the name of the application to be deployed to the Cloud Foundry space.
         * @parentConfigKey cloudFoundry
         */
        'appName',
        /**
         * Credentials to be used for deployment.
         * @parentConfigKey cloudFoundry
         */
        'credentialsId',
        /**
         * Defines the manifest to be used for deployment to Cloud Foundry.
         * @parentConfigKey cloudFoundry
         */
        'manifest',
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
     * Defines the tool which should be used for deployment.
     * @possibleValues 'cf_native', 'mtaDeployPlugin'
     */
    'deployTool',
    /**
     * Defines the type of deployment, either `standard` deployment which results in a system downtime or a zero-downtime `blue-green` deployment.
     * @possibleValues 'standard', 'blue-green'
     */
    'deployType',
    /**
     * In case of a `blue-green` deployment the old instance will be deleted by default. If this option is set to true the old instance will remain stopped in the Cloud Foundry space.
     * @possibleValues true, false
     */
    'keepOldInstance',
    /** @see dockerExecute */
    'dockerImage',
    /** @see dockerExecute */
    'dockerWorkspace',
    /** @see dockerExecute */
    'stashContent',
    /**
     * Defines additional parameters passed to mta for deployment with the mtaDeployPlugin.
     */
    'mtaDeployParameters',
    /**
     * Defines additional extension descriptor file for deployment with the mtaDeployPlugin.
     */
    'mtaExtensionDescriptor',
    /**
     * Defines the path to *.mtar for deployment with the mtaDeployPlugin.
     */
    'mtaPath',
    /**
     * Allows to specify a script which performs a check during blue-green deployment. The script gets the FQDN as parameter and returns `exit code 0` in case check returned `smokeTestStatusCode`.
     * More details can be found [here](https://github.com/bluemixgaragelondon/cf-blue-green-deploy#how-to-use) <br /> Currently this option is only considered for deployTool `cf_native`.
     */
    'smokeTestScript',
    /**
     * Expected status code returned by the check.
     */
    'smokeTestStatusCode'
]

@Field Map CONFIG_KEY_COMPATIBILITY = [cloudFoundry: [apiEndpoint: 'cfApiEndpoint', appName:'cfAppName', credentialsId: 'cfCredentialsId', manifest: 'cfManifest', org: 'cfOrg', space: 'cfSpace']]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Deploys an application to a test or production space within Cloud Foundry.
 * Deployment can be done
 *
 * * in a standard way
 * * in a zero downtime manner (using a [blue-green deployment approach](https://martinfowler.com/bliki/BlueGreenDeployment.html))
 *
 * !!! note "Deployment supports multiple deployment tools"
 *     Currently the following are supported:
 *
 *     * Standard `cf push` and [Bluemix blue-green plugin](https://github.com/bluemixgaragelondon/cf-blue-green-deploy#how-to-use)
 *     * [MTA CF CLI Plugin](https://github.com/cloudfoundry-incubator/multiapps-cli-plugin)
 *
 * !!! note
 * Due to [an incompatible change](https://github.com/cloudfoundry/cli/issues/1445) in the Cloud Foundry CLI, multiple buildpacks are not supported by this step.
 * If your `application` contains a list of `buildpacks` instead a single `buildpack`, this will be automatically re-written by the step when blue-green deployment is used.
 *
 * !!! note
 * Cloud Foundry supports the deployment of multiple applications using a single manifest file.
 * This option is supported with Piper.
 *
 * In this case define `appName: ''` since the app name for the individual applications have to be defined via the manifest.
 * You can find details in the [Cloud Foundry Documentation](https://docs.cloudfoundry.org/devguide/deploy-apps/manifest.html#multi-apps)
 */
@GenerateDocumentation
void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def utils = parameters.juStabUtils ?: new Utils()
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()

        final script = checkScript(this, parameters) ?: this

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

        echo "[${STEP_NAME}] General parameters: deployTool=${config.deployTool}, deployType=${config.deployType}, cfApiEndpoint=${config.cloudFoundry.apiEndpoint}, cfOrg=${config.cloudFoundry.org}, cfSpace=${config.cloudFoundry.space}, cfCredentialsId=${config.cloudFoundry.credentialsId}"

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
    def mtarFiles = findFiles(glob: '**/*.mtar')

    if(mtarFiles.length > 1){
        error "Found multiple *.mtar files, please specify file via mtaPath parameter! ${mtarFiles}"
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

        Exception exceptionFromDeploy, exceptionFromPostDeploy

        try {
            sh script: """#!/bin/bash
                set +x
                set -e
                export HOME=${config.dockerWorkspace}
                cf login -u \"${username}\" -p '${password}' -a ${config.cloudFoundry.apiEndpoint} -o \"${config.cloudFoundry.org}\" -s \"${config.cloudFoundry.space}\"
                cf plugins
                cf ${deployCommand} ${config.cloudFoundry.appName ?: ''} ${blueGreenDeployOptions} -f '${config.cloudFoundry.manifest}' ${config.smokeTest}
            """
        } catch(Exception e) {
            exceptionFromDeploy = e
        } finally {
            try {
                stopOldAppIfRunning(config)
                sh "cf logout"
            } catch(Exception e) {
                exceptionFromPostDeploy = e
            }
        }
        if(exceptionFromDeploy) {
            if(exceptionFromPostDeploy) {
                // [Q] What is the reason for the echo statement below?
                // [A] In case the exceptionFromDeploy is a hudson.AbortException we only see the message from the exception in the log - no stacktrace.
                //     Hence the suppressed exception - which has been in fact added to the hudson.AbortException is not visible.
                echo "Got an exception from the deployment step and another exception from the cleanup. The exception from cleanup was: '${exceptionFromPostDeploy}'."
                exceptionFromDeploy.addSuppressed(exceptionFromPostDeploy)
            }
            throw exceptionFromDeploy
        }
        if(exceptionFromPostDeploy) {
            throw exceptionFromPostDeploy
        }
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

private void stopOldAppIfRunning(Map config) {
    String oldAppName = "${config.cloudFoundry.appName}-old"
    String cfStopOutputFileName = "${UUID.randomUUID()}-cfStopOutput.txt"

    if (config.keepOldInstance && config.deployType == 'blue-green') {
        int cfStopReturncode = sh (returnStatus: true, script: "cf stop $oldAppName  &> $cfStopOutputFileName")

        if (cfStopReturncode > 0) {
            String cfStopOutput = readFile(file: cfStopOutputFileName)

            if (!cfStopOutput.contains("$oldAppName not found")) {
                error "Could not stop application $oldAppName. Error: $cfStopOutput"
            }
        }
    }
}

def deployMta (config) {
    if (config.mtaExtensionDescriptor == null) config.mtaExtensionDescriptor = ''
    if (!config.mtaExtensionDescriptor.isEmpty() && !config.mtaExtensionDescriptor.startsWith('-e ')) config.mtaExtensionDescriptor = "-e ${config.mtaExtensionDescriptor}"

    def deployCommand = 'deploy'
    if (config.deployType == 'blue-green') {
        deployCommand = 'bg-deploy'
        if (config.mtaDeployParameters.indexOf('--no-confirm') < 0) {
            config.mtaDeployParameters += ' --no-confirm'
        }
    }

    withCredentials([usernamePassword(
        credentialsId: config.cloudFoundry.credentialsId,
        passwordVariable: 'password',
        usernameVariable: 'username'
    )]) {
        echo "[${STEP_NAME}] Deploying MTA (${config.mtaPath}) with following parameters: ${config.mtaExtensionDescriptor} ${config.mtaDeployParameters}"

        Exception exceptionFromDeploy, exceptionFromPostDeploy

        try {
            sh returnStatus: true, script: """#!/bin/bash
                export HOME=${config.dockerWorkspace}
                set +x
                set -e
                cf api ${config.cloudFoundry.apiEndpoint}
                cf login -u ${username} -p '${password}' -a ${config.cloudFoundry.apiEndpoint} -o \"${config.cloudFoundry.org}\" -s \"${config.cloudFoundry.space}\"
                cf plugins
                cf ${deployCommand} ${config.mtaPath} ${config.mtaDeployParameters} ${config.mtaExtensionDescriptor}"""
        } catch(Exception e) {
            exceptionFromDeploy = e
        } finally {
            try {
                sh "cf logout"
            } catch(Exception e) {
                exceptionFromPostDeploy
            }
        }

        if(exceptionFromDeploy) {
            if(exceptionFromPostDeploy) {
                echo "Exception caught during post deploy action: ${exceptionFromPostDeploy}"
                exceptionFromDeploy.addSuppressed(exceptionFromPostDeploy)
            }
            throw exceptionFromDeploy
        }

        if(exceptionFromPostDeploy) {
            throw exceptionFromPostDeploy
        }
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
