import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.JenkinsUtils
import com.sap.piper.JsonUtils
import com.sap.piper.Utils
import com.sap.piper.integration.TransportManagementService
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/tmsUpload.yaml'

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
     * Defines the relative path to *.mtar for the upload to the Transport Management Service. If not specified, it will use the mtar file created in mtaBuild.
     */
    'mtaPath',
    /**
     * Defines the name of the node to which the *.mtar file should be uploaded.
     */
    'nodeName',
    /**
     * Defines the version of the MTA for which the MTA extension descriptor will be used. You can use an asterisk (*) to accept any MTA version, or use a specific version compliant with SemVer 2.0, e.g. 1.0.0 (see semver.org). If the parameter is not configured, an asterisk is used.
     */
    'mtaVersion',
    /**
     * Available only for transports in Cloud Foundry environment. Defines a mapping between a transport node name and an MTA extension descriptor file path that you want to use for the transport node, e.g. nodeExtDescriptorMapping: [nodeName: 'example.mtaext', nodeName2: 'example2.mtaext', …]`.
     */
    'nodeExtDescriptorMapping',
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
    'proxy',
    /**
     * The new Golang implementation of the step is used now by default. Utilizing this toggle with value true is therefore redundant and can be omitted. If used with value false, the toggle deactivates the new Golang implementation and instructs the step to use the old Groovy one. Note that possibility to switch to the old Groovy implementation will be completely removed and this toggle will be deprecated after February 29th, 2024.
     * @possibleValues true, false
     */
    'useGoStep'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS + GENERAL_CONFIG_KEYS

/**
 * This step allows you to upload an MTA file (multi-target application archive) and multiple MTA extension descriptors into a TMS (SAP BTP Transport Management Service) landscape for further TMS-controlled distribution through a TMS-configured landscape.
 * TMS lets you manage transports between SAP BTP accounts in Neo and Cloud Foundry, such as from DEV to TEST and PROD accounts.
 * For more information, see [official documentation of Transport Management Service](https://help.sap.com/viewer/p/TRANSPORT_MANAGEMENT_SERVICE)
 *
 * !!! note "Prerequisites"
 *     * You have subscribed to and set up TMS, as described in [Setup and Configuration of SAP BTP Transport Management](https://help.sap.com/viewer/7f7160ec0d8546c6b3eab72fb5ad6fd8/Cloud/en-US/66fd7283c62f48adb23c56fb48c84a60.html), which includes the configuration of a node to be used for uploading an MTA file.
 *     * A corresponding service key has been created, as described in [Set Up the Environment to Transport Content Archives directly in an Application](https://help.sap.com/viewer/7f7160ec0d8546c6b3eab72fb5ad6fd8/Cloud/en-US/8d9490792ed14f1bbf8a6ac08a6bca64.html). This service key (JSON) must be stored as a secret text within the Jenkins secure store.
 *
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()
        String stageName = parameters.stageName ?: env.STAGE_NAME

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .addIfEmpty('mtaPath', script.commonPipelineEnvironment.mtarFilePath)
            //mandatory parameters
            .withMandatoryProperty('mtaPath')
            .withMandatoryProperty('nodeName')
            .withMandatoryProperty('credentialsId')
            .use()

        def namedUser = jenkinsUtils.getJobStartedByUserId()

        if (config.useGoStep != false) {
            List credentials = [
                [type: 'token', id: 'credentialsId', env: ['PIPER_serviceKey']]
            ]

            if (namedUser) {
                parameters.namedUser = namedUser
            }

            utils.unstashAll(config.stashContent)
            piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
            return
        }

        echo "[TransportManagementService] Using deprecated Groovy implementation of '${STEP_NAME}' step instead of the default Golang one, since 'useGoStep' toggle parameter is explicitly set to 'false'."
        echo "[TransportManagementService] WARNING: Note that the deprecated Groovy implementation will be completely removed after February 29th, 2024. Consider using the Golang implementation by not setting the 'useGoStep' toggle parameter to 'false'."

        def jsonUtilsObject = new JsonUtils()

        // make sure that all relevant descriptors, are available in workspace
        utils.unstashAll(config.stashContent)
        // make sure that for further execution whole workspace, e.g. also downloaded artifacts are considered
        config.stashContent = []

        def customDescription = config.customDescription ? "${config.customDescription}" : "Git CommitId: ${script.commonPipelineEnvironment.getGitCommitId()}"
        def description = customDescription

        if (!namedUser) {
            namedUser = config.namedUser
        }

        def nodeName = config.nodeName
        def mtaPath = config.mtaPath

        def mtaVersion = config.mtaVersion ? "${config.mtaVersion}" : "*"
        Map nodeExtDescriptorMapping = (config.nodeExtDescriptorMapping && config.nodeExtDescriptorMapping.size()>0) ? config.nodeExtDescriptorMapping : null

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

            if(nodeExtDescriptorMapping) {
                // validate the whole mapping and then throw errors together,
                // so that user can get them in one pipeline run
                // put the validation here, because we need uri and token to call tms get nodes api
                List nodes = tms.getNodes(uri, token).getAt("nodes");
                Map mtaYaml = getMtaYaml(script.commonPipelineEnvironment.getValue('mtaBuildToolDesc'));
                Map nodeIdExtDesMap = validateNodeExtDescriptorMapping(nodeExtDescriptorMapping, nodes, mtaYaml, mtaVersion)

                if(nodeIdExtDesMap) {
                    nodeIdExtDesMap.each{ key, value ->
                        Map mtaExtDescriptor = tms.getMtaExtDescriptor(uri, token, key, mtaYaml.ID, mtaVersion)
                        if(mtaExtDescriptor) {
                            def updateMtaExtDescriptorResponse = tms.updateMtaExtDescriptor(uri, token, key, mtaExtDescriptor.getAt("id"), "${workspace}/${value.get(1)}", mtaVersion, description, namedUser)
                            echo "[TransportManagementService] MTA Extension Descriptor with ID '${updateMtaExtDescriptorResponse.mtaExtId}' successfully updated for Node '${value.get(0)}'."
                        } else {
                            def uploadMtaExtDescriptorToNodeResponse = tms.uploadMtaExtDescriptorToNode(uri, token, key, "${workspace}/${value.get(1)}", mtaVersion, description, namedUser)
                            echo "[TransportManagementService] MTA Extension Descriptor with ID '${uploadMtaExtDescriptorToNodeResponse.mtaExtId}' successfully uploaded to Node '${value.get(0)}'."
                        }
                    }
                }
            }

            def fileUploadResponse = tms.uploadFile(uri, token, "${workspace}/${mtaPath}", namedUser)
            def uploadFileToNodeResponse = tms.uploadFileToNode(uri, token, nodeName, fileUploadResponse.fileId, description, namedUser)
            echo "[TransportManagementService] File '${fileUploadResponse.fileName}' successfully uploaded to Node '${uploadFileToNodeResponse.queueEntries.nodeName}' (Id: '${uploadFileToNodeResponse.queueEntries.nodeId}')."
            echo "[TransportManagementService] Corresponding Transport Request: '${uploadFileToNodeResponse.transportRequestDescription}' (Id: '${uploadFileToNodeResponse.transportRequestId}')"
        }

    }
}

def String getMtaId(String extDescriptorFilePath){
    def extDescriptor = readYaml file: extDescriptorFilePath
    def mtaId = ""
    if (extDescriptor.extends) {
        mtaId = extDescriptor.extends
    }
    return mtaId
}

def Map getMtaYaml(String mtaBuildToolDesc) {
    mtaBuildToolDesc = mtaBuildToolDesc?:"mta.yaml"
    if(fileExists(mtaBuildToolDesc)) {
        def mtaYaml = readYaml file: mtaBuildToolDesc
        if (!mtaYaml.ID || !mtaYaml.version) {
            def errorMsg
            if (!mtaYaml.ID) {
                errorMsg = "Property 'ID' is not found in ${mtaBuildToolDesc}."
            }
            if (!mtaYaml.version) {
                errorMsg += "Property 'version' is not found in ${mtaBuildToolDesc}."
            }
            error errorMsg
        }
        return mtaYaml
    } else {
        error "${mtaBuildToolDesc} is not found in the root folder of the project."
    }
}

def Map validateNodeExtDescriptorMapping(Map nodeExtDescriptorMapping, List nodes, Map mtaYaml, String mtaVersion) {
    def errorPathList = []
    def errorMtaId = []
    def errorNodeNameList = []
    def errorMsg = ""
    Map nodeIdExtDesMap = [:]

    if(mtaVersion != "*" && mtaVersion != mtaYaml.version) {
        errorMsg = "Parameter 'mtaVersion' does not match the MTA version in mta.yaml. "
    }

    nodeExtDescriptorMapping.each{ key, value ->
        if(nodes.any {it.name == key}) {
            nodeIdExtDesMap.put(nodes.find {it.name == key}.getAt("id"), [key, value])
        } else {
            errorNodeNameList.add(key)
        }

        if(!fileExists(value)) {
            errorPathList.add(value)
        } else {
            if(mtaYaml.ID != getMtaId("${value}")) {
                errorMtaId.add(value)
            }
        }
    }

    if(!errorPathList.isEmpty() || !errorMtaId.isEmpty() || !errorNodeNameList.isEmpty() ) {
        if(!errorPathList.isEmpty()) {
            errorMsg += "MTA extension descriptor files ${errorPathList} don't exist. "
        }
        if(!errorMtaId.isEmpty()) {
            errorMsg += "Parameter [extends] in MTA extension descriptor files ${errorMtaId} is not the same as MTA ID."
        }
        if(!errorNodeNameList.isEmpty()) {
            errorMsg += "Nodes ${errorNodeNameList} don't exist. Please check the node name or create these nodes."
        }
        error(errorMsg)
    }

    return nodeIdExtDesMap
}
