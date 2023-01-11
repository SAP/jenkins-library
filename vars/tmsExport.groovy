import groovy.transform.Field
import com.sap.piper.JenkinsUtils

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/tmsExport.yaml'
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
    'proxy'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS + GENERAL_CONFIG_KEYS

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'token', id: 'credentialsId', env: ['PIPER_tmsServiceKey']]
    ]

    if (!parameters.namedUser) {
        def jenkinsUtils = new JenkinsUtils()
        def namedUser = jenkinsUtils.getJobStartedByUserId()
        if (namedUser) {
            parameters.namedUser = namedUser
        }
    }    
    
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials, false, false, true)
}