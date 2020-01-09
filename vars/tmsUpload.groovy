import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.JenkinsUtils
import com.sap.piper.JsonUtils
import com.sap.piper.Utils
import com.sap.piper.integration.TransportManagementService
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Print more detailed information into the log.
     * @possibleValues `true`, `false`
     */
    'verbose'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * If specific stashes should be considered, their names need to be passed via the parameter `stashContent`.
     */
    'stashContent',
    /**
     * Defines the path to *.mtar for the upload to the Transport Management Service.
     */
    'mtaPath',
    /**
     * Defines the name of the node to which the *.mtar file should be uploaded.
     */
    'nodeName',
    /**
     * Credentials to be used for the file and node uploads to the Transport Management Service.
     */
    'credentialsId',
    /**
     * Can be used as the description of a transport request. Will overwrite the default. (Default: Corresponding Git Commit-ID)
     */
    'customDescription',
    /**
     * Proxy which should be used for the communication with the Transport Management Service Backend.
     */
    'proxy'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * This step allows you to upload an MTA file (multi-target application archive) into a TMS (SAP Cloud Platform Transport Management Service) landscape for further TMS-controlled distribution through a TMS-configured landscape.
 * TMS lets you manage transports between SAP Cloud Platform accounts in Neo and Cloud Foundry, such as from DEV to TEST and PROD accounts.
 * For more information, see [official documentation of Transport Management Service](https://help.sap.com/viewer/p/TRANSPORT_MANAGEMENT_SERVICE)
 *
 * !!! note "Prerequisites"
 *     * You have subscribed to and set up TMS, as described in [Setup and Configuration of SAP Cloud Platform Transport Management](https://help.sap.com/viewer/7f7160ec0d8546c6b3eab72fb5ad6fd8/Cloud/en-US/66fd7283c62f48adb23c56fb48c84a60.html), which includes the configuration of a node to be used for uploading an MTA file.
 *     * A corresponding service key has been created, as described in [Set Up the Environment to Transport Content Archives directly in an Application](https://help.sap.com/viewer/7f7160ec0d8546c6b3eab72fb5ad6fd8/Cloud/en-US/8d9490792ed14f1bbf8a6ac08a6bca64.html). This service key (JSON) must be stored as a secret text within the Jenkins secure store.
 *
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this, script)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName ?: env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            //mandatory parameters
            .withMandatoryProperty('mtaPath')
            .withMandatoryProperty('nodeName')
            .withMandatoryProperty('credentialsId')
            .use()

        // telemetry reporting
        new Utils().pushToSWA([
            step         : STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1   : parameters?.script == null
        ], config)

        def jsonUtilsObject = new JsonUtils()

        // make sure that all relevant descriptors, are available in workspace
        utils.unstashAll(config.stashContent)
        // make sure that for further execution whole workspace, e.g. also downloaded artifacts are considered
        config.stashContent = []

        def customDescription = config.customDescription ? "${config.customDescription}" : "Git CommitId: ${script.commonPipelineEnvironment.getGitCommitId()}"
        def description = customDescription

        def namedUser = jenkinsUtils.getJobStartedByUserId() ?: config.namedUser

        def nodeName = config.nodeName
        def mtaPath = config.mtaPath

        if(!fileExists(mtaPath)) {
            error("Mta file '${mtaPath}' does not exist.")
        }

        if (config.verbose) {
            echo "[TransportManagementService] CredentialsId: '${config.credentialsId}'"
            echo "[TransportManagementService] Node name: '${nodeName}'"
            echo "[TransportManagementService] MTA path: '${mtaPath}'"
            echo "[TransportManagementService] Named user: '${namedUser}'"
        }

        def tms = parameters.transportManagementService ?: new TransportManagementService(script, config)

        withCredentials([string(credentialsId: config.credentialsId, variable: 'tmsServiceKeyJSON')]) {

            def tmsServiceKey = jsonUtilsObject.jsonStringToGroovyObject(tmsServiceKeyJSON)

            def clientId = tmsServiceKey.uaa.clientid
            def clientSecret = tmsServiceKey.uaa.clientsecret
            def uaaUrl = tmsServiceKey.uaa.url
            def uri = tmsServiceKey.uri

            if (config.verbose) {
                echo "[TransportManagementService] UAA URL: '${uaaUrl}'"
                echo "[TransportManagementService] TMS URL: '${uri}'"
                echo "[TransportManagementService] ClientId: '${clientId}'"
            }

            def token = tms.authentication(uaaUrl, clientId, clientSecret)
            def fileUploadResponse = tms.uploadFile(uri, token, "${workspace}/${mtaPath}", namedUser)
            def uploadFileToNodeResponse = tms.uploadFileToNode(uri, token, nodeName, fileUploadResponse.fileId, description, namedUser)

            echo "[TransportManagementService] File '${fileUploadResponse.fileName}' successfully uploaded to Node '${uploadFileToNodeResponse.queueEntries.nodeName}' (Id: '${uploadFileToNodeResponse.queueEntries.nodeId}')."
            echo "[TransportManagementService] Corresponding Transport Request: '${uploadFileToNodeResponse.transportRequestDescription}' (Id: '${uploadFileToNodeResponse.transportRequestId}')"

        }

    }
}
