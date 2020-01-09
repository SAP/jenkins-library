import com.sap.piper.JenkinsUtils

import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
import com.sap.piper.CfManifestUtils
import com.sap.piper.BashUtils

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
         * Defines the manifest variables Yaml files to be used to replace variable references in manifest. This parameter
         * is optional and will default to `["manifest-variables.yml"]`. This can be used to set variable files like it
         * is provided by `cf push --vars-file <file>`.
         *
         * If the manifest is present and so are all variable files, a variable substitution will be triggered that uses
         * the `cfManifestSubstituteVariables` step before deployment. The format of variable references follows the
         * [Cloud Foundry standard](https://docs.cloudfoundry.org/devguide/deploy-apps/manifest-attributes.html#variable-substitution).
         * @parentConfigKey cloudFoundry
         */
        'manifestVariablesFiles',
        /**
         * Defines a `List` of variables as key-value `Map` objects used for variable substitution within the file given by `manifest`.
         * Defaults to an empty list, if not specified otherwise. This can be used to set variables like it is provided
         * by `cf push --var key=value`.
         *
         * The order of the maps of variables given in the list is relevant in case there are conflicting variable names and values
         * between maps contained within the list. In case of conflicts, the last specified map in the list will win.
         *
         * Though each map entry in the list can contain more than one key-value pair for variable substitution, it is recommended
         * to stick to one entry per map, and rather declare more maps within the list. The reason is that
         * if a map in the list contains more than one key-value entry, and the entries are conflicting, the
         * conflict resolution behavior is undefined (since map entries have no sequence).
         *
         * Note: variables defined via `manifestVariables` always win over conflicting variables defined via any file given
         * by `manifestVariablesFiles` - no matter what is declared before. This is the same behavior as can be
         * observed when using `cf push --var` in combination with `cf push --vars-file`.
         */
        'manifestVariables',
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
     * Additional parameters passed to cf native deployment command.
     */
    'cfNativeDeployParameters',
    /**
     * Addition command line options for cf api command.
     * No escaping/quoting is performed. Not recommanded for productive environments.
     */
    'apiParameters',
    /**
     * Addition command line options for cf login command.
     * No escaping/quoting is performed. Not recommanded for productive environments.
     */
    'loginParameters',
    /**
     * Additional parameters passed to mta deployment command.
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
    'smokeTestStatusCode',
    /**
      * Provides more output. May reveal sensitive information.
      * @possibleValues true, false
      */
    'verbose',
]

@Field Map CONFIG_KEY_COMPATIBILITY = [cloudFoundry: [apiEndpoint: 'cfApiEndpoint', appName:'cfAppName', credentialsId: 'cfCredentialsId', manifest: 'cfManifest', manifestVariablesFiles: 'cfManifestVariablesFiles', manifestVariables: 'cfManifestVariables',  org: 'cfOrg', space: 'cfSpace']]

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

        Map config = ConfigurationHelper.newInstance(this, script)
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

        boolean deployTriggered = false
        boolean deploySuccess = true
        try {
            if (config.deployTool == 'mtaDeployPlugin') {
                deployTriggered = true
                handleMTADeployment(config, script)
            }
            else if (config.deployTool == 'cf_native') {
                deployTriggered = true
                handleCFNativeDeployment(config, script)
            }
            else {
                deployTriggered = false
                echo "[${STEP_NAME}] WARNING! Found unsupported deployTool. Skipping deployment."
            }
        } catch (err) {
            deploySuccess = false
            throw err
        } finally {
            if (deployTriggered) {
                reportToInflux(script, config, deploySuccess, jenkinsUtils)
            }

        }

    }
}

private void handleMTADeployment(Map config, script) {
    // set default mtar path
    config = ConfigurationHelper.newInstance(this, script, config)
        .addIfEmpty('mtaPath', config.mtaPath ?: findMtar())
        .use()

    dockerExecute(script: script, dockerImage: config.dockerImage, dockerWorkspace: config.dockerWorkspace, stashContent: config.stashContent) {
        deployMta(config)
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

    def deployStatement = "cf ${deployCommand} ${config.mtaPath} ${config.mtaDeployParameters} ${config.mtaExtensionDescriptor}"
    def apiStatement = "cf api ${config.cloudFoundry.apiEndpoint} ${config.apiParameters}"

    echo "[${STEP_NAME}] Deploying MTA (${config.mtaPath}) with following parameters: ${config.mtaExtensionDescriptor} ${config.mtaDeployParameters}"
    deploy(apiStatement, deployStatement, config, null)
}

private void handleCFNativeDeployment(Map config, script) {
    config.smokeTest = ''

    if (config.deployType == 'blue-green') {
        prepareBlueGreenCfNativeDeploy(config,script)
    } else {
        prepareCfPushCfNativeDeploy(config)
    }

    echo "[${STEP_NAME}] CF native deployment (${config.deployType}) with:"
    echo "[${STEP_NAME}] - cfAppName=${config.cloudFoundry.appName}"
    echo "[${STEP_NAME}] - cfManifest=${config.cloudFoundry.manifest}"
    echo "[${STEP_NAME}] - cfManifestVariables=${config.cloudFoundry.manifestVariables?:'none specified'}"
    echo "[${STEP_NAME}] - cfManifestVariablesFiles=${config.cloudFoundry.manifestVariablesFiles?:'none specified'}"
    echo "[${STEP_NAME}] - smokeTestScript=${config.smokeTestScript}"

    checkIfAppNameIsAvailable(config)
    dockerExecute(
        script: script,
        dockerImage: config.dockerImage,
        dockerWorkspace: config.dockerWorkspace,
        stashContent: config.stashContent,
        dockerEnvVars: [CF_HOME: "${config.dockerWorkspace}", CF_PLUGIN_HOME: "${config.dockerWorkspace}", STATUS_CODE: "${config.smokeTestStatusCode}"]
    ) {
        deployCfNative(config)
    }
}

private prepareBlueGreenCfNativeDeploy(config,script) {
    if (config.smokeTestScript == 'blueGreenCheckScript.sh') {
        writeFile file: config.smokeTestScript, text: libraryResource(config.smokeTestScript)
    }

    config.smokeTest = '--smoke-test $(pwd)/' + config.smokeTestScript
    sh "chmod +x ${config.smokeTestScript}"

    config.deployCommand = 'blue-green-deploy'
    cfManifestSubstituteVariables(
        script: script,
        manifestFile: config.cloudFoundry.manifest,
        manifestVariablesFiles: config.cloudFoundry.manifestVariablesFiles,
        manifestVariables: config.cloudFoundry.manifestVariables
    )
    handleLegacyCfManifest(config)
    if (!config.keepOldInstance) {
        config.deployOptions = '--delete-old-apps'
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

private prepareCfPushCfNativeDeploy(config) {
    config.deployCommand = 'push'
    config.deployOptions = "${varOptions(config)}${varFileOptions(config)}"
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

private checkIfAppNameIsAvailable(config) {
    if (config.cloudFoundry.appName == null || config.cloudFoundry.appName == '') {
        if (config.deployType == 'blue-green') {
            error "[${STEP_NAME}] ERROR: Blue-green plugin requires app name to be passed (see https://github.com/bluemixgaragelondon/cf-blue-green-deploy/issues/27)"
        }
        if (fileExists(config.cloudFoundry.manifest)) {
            def manifest = readYaml file: config.cloudFoundry.manifest
            if (!manifest || !manifest.applications || !manifest.applications[0].name) {
                error "[${STEP_NAME}] ERROR: No appName available in manifest ${config.cloudFoundry.manifest}."
            }
        } else {
            error "[${STEP_NAME}] ERROR: No manifest file ${config.cloudFoundry.manifest} found."
        }
    }
}

def deployCfNative (config) {
    def deployStatement = "cf ${config.deployCommand} ${config.cloudFoundry.appName ?: ''} ${config.deployOptions?:''} -f '${config.cloudFoundry.manifest}' ${config.smokeTest} ${config.cfNativeDeployParameters}"
    deploy(null, deployStatement, config,  { c -> stopOldAppIfRunning(c) })
}

private deploy(def cfApiStatement, def cfDeployStatement, def config, Closure postDeployAction) {

    withCredentials([usernamePassword(
        credentialsId: config.cloudFoundry.credentialsId,
        passwordVariable: 'password',
        usernameVariable: 'username'
    )]) {

        def cfTraceFile = 'cf.log'

        def deployScript = """#!/bin/bash
            set +x
            set -e
            export HOME=${config.dockerWorkspace}
            export CF_TRACE=${cfTraceFile}
            ${cfApiStatement ?: ''}
            cf login -u \"${username}\" -p '${password}' -a ${config.cloudFoundry.apiEndpoint} -o \"${config.cloudFoundry.org}\" -s \"${config.cloudFoundry.space}\" ${config.loginParameters}
            cf plugins
            ${cfDeployStatement}
            """

        if(config.verbose) {
            // Password contained in output below is hidden by withCredentials
            echo "[INFO][${STEP_NAME}] Executing command: '${deployScript}'."
        }

        def returnCode = sh returnStatus: true, script: deployScript

        if(config.verbose || returnCode != 0) {
            if(fileExists(file: cfTraceFile)) {
                echo  '### START OF CF CLI TRACE OUTPUT ###'
                // Would be nice to inline the two next lines, but that is not understood by the test framework
                def cfTrace =  readFile(file: cfTraceFile)
                echo cfTrace
                echo '### END OF CF CLI TRACE OUTPUT ###'
            } else {
                echo "No trace file found at '${cfTraceFile}'"
            }
        }

        if(returnCode != 0){
            error "[${STEP_NAME}] ERROR: The execution of the deploy command failed, see the log for details."
        }

        if(postDeployAction) postDeployAction(config)

        sh "cf logout"
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
                error "[${STEP_NAME}] ERROR: Could not stop application $oldAppName. Error: $cfStopOutput"
            }
        }
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
