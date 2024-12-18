import com.sap.piper.Utils
import groovy.transform.Field
import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/cloudFoundryDeploy.yaml'

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    String stageName = parameters.stageName ?: env.STAGE_NAME

    def utils = parameters.juStabUtils ?: new Utils()
    utils.unstashAll(["deployDescriptor"])
    List credentials = [
        [type: 'usernamePassword', id: 'cfCredentialsId', env: ['PIPER_username', 'PIPER_password']],
        [type: 'usernamePassword', id: 'dockerCredentialsId', env: ['PIPER_dockerUsername', 'PIPER_dockerPassword']]
    ]

    Map mtaExtensionCredentials = parameters.mtaExtensionCredentials ?: script.commonPipelineEnvironment.getStepConfiguration(STEP_NAME, stageName).mtaExtensionCredentials
    Bool checkMissingCredentials = parameters.checkMissingCredentials ?: script.commonPipelineEnvironment.getStepConfiguration(STEP_NAME, stageName).checkMissingCredentials

    if (mtaExtensionCredentials) {
        if (checkMissingCredentials) {
            mtaExtensionCredentials.each { key, credentialsId ->
            echo "[INFO]${STEP_NAME}] Preparing credential for being used by piper-go. key: ${key}, credentialsId is: ${credentialsId}, exposed as environment variable ${toEnvVarKey(credentialsId)}"
            credentials << [type: 'token', id: credentialsId, env: [toEnvVarKey(credentialsId)], resolveCredentialsId: false]
            }
        }
    }
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}

/*
 * Inserts underscores before all upper case letters which are not already
 * have an underscore before, replaces any non letters/digits with underscore
 * and transforms all lower case letters to upper case.
 */
private static String toEnvVarKey(String key) {
    key = key.replaceAll(/[^A-Za-z0-9]/, "_")
    key = key.replaceAll(/([a-z0-9])([A-Z])/, /$1_$2/)
    return key.toUpperCase()
}
