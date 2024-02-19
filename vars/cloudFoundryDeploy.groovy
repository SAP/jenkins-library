import com.sap.piper.Utils
import groovy.transform.Field
import org.jenkinsci.plugins.plaincredentials.StringCredentials

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/cloudFoundryDeploy.yaml'

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()
    String stageName = parameters.stageName ?: env.STAGE_NAME
    String piperGoPath = parameters.piperGoPath ?: './piper'

    // Set the default credential type to usernamePassword
    def cfCred = [type: 'usernamePassword', id: 'cfCredentialsId', env: ['PIPER_username', 'PIPER_password']]

    withEnv([
        "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(stepParameters)}",
        "PIPER_correlationID=${env.BUILD_URL}",
    ]) {
        String customDefaultConfig = piperExecuteBin.getCustomDefaultConfigsArg()
        String customConfigArg = piperExecuteBin.getCustomConfigArg(script)
        Map config
        piperExecuteBin.handleErrorDetails(STEP_NAME) {
            config = piperExecuteBin.getStepContextConfig(script, piperGoPath, METADATA_FILE, customDefaultConfig, customConfigArg)
            echo "Context Config: ${config}"
        }
        def cfCredentialsId = config['cfCredentialsId']
        if (cfCredentialsId) {
            def cfCredential = utils.getJenkinsCredentialEntry(cfCredentialsId)
            if (cfCredential instanceof StringCredentials) {
                cfCred = [type: 'token', id: 'cfCredentialsId', env: ['PIPER_IDP_JSON_CREDENTIAL']]
            }
        }
    }
    utils.unstashAll(["deployDescriptor"])
    List credentials = [
        cfCred,
        [type: 'usernamePassword', id: 'dockerCredentialsId', env: ['PIPER_dockerUsername', 'PIPER_dockerPassword']]
    ]

    Map mtaExtensionCredentials = parameters.mtaExtensionCredentials ?: script.commonPipelineEnvironment.getStepConfiguration(STEP_NAME, stageName).mtaExtensionCredentials

    if (mtaExtensionCredentials) {
        mtaExtensionCredentials.each { key, credentialsId ->
            echo "[INFO]${STEP_NAME}] Preparing credential for being used by piper-go. key: ${key}, credentialsId is: ${credentialsId}, exposed as environment variable ${toEnvVarKey(credentialsId)}"
            credentials << [type: 'token', id: credentialsId, env: [toEnvVarKey(credentialsId)], resolveCredentialsId: false]
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
