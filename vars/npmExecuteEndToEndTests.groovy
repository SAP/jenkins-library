import com.sap.piper.ConfigurationHelper
import com.sap.piper.DownloadCacheUtils
import com.sap.piper.GenerateDocumentation
import com.sap.piper.k8s.ContainerMap
import groovy.transform.Field
import com.sap.piper.Utils

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/npmExecuteEndToEndTests.yaml'

@Field Set STEP_CONFIG_KEYS = [
    /**
     * The URLs under which the app is available after deployment.
     * Each element of appUrls must be a map containing a property url, an optional property credentialId, and an optional property parameters.
     * The optional property parameters can be used to pass additional parameters to the end-to-end test deployment reachable via the given application URL.
     * These parameters must be a list of strings, where each string corresponds to one element of the parameters.
     * For example, if the parameter `--tag scenario1` should be passed to the test, specify parameters: ["--tag", "scenario1"].
     * These parameters are appended to the npm command during execution.
     */
    'appUrls',
    /**
     * Credentials to access the application to be tested
     */
    'credentialsId',
    /**
     * Distinguish if these are wdi5 tests. If set to `true` `wdi5_username` and `wdi5_password` environment variables are used to enable [autologin](https://ui5-community.github.io/wdi5/#/authentication?id=credentials).
     * @possibleValues `true`, `false`
     */
    'wdi5'
]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

@Field Map CONFIG_KEY_COMPATIBILITY = [parallelExecution: 'features/parallelTestExecution']

void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    String stageName = parameters.stageName ?: env.STAGE_NAME
    Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

    if (config.appUrls && !(config.appUrls instanceof List)) {
        error "[${STEP_NAME}] The execution failed, since appUrls is not a list. Please provide appUrls as a list of maps. For example:\n" +
                "appUrls: \n" + "  - url: 'https://my-url.com'\n" + "    credentialId: myCreds"
    }
    List credentials = []
    if (config.appUrls){
        for (int i = 0; i < config.appUrls.size(); i++) {
            def appUrl = config.appUrls[i]
            if (!(appUrl instanceof Map)) {
                error "[${STEP_NAME}] The element ${appUrl} is not of type map. Please provide appUrls as a list of maps. For example:\n" +
                        "appUrls: \n" + "  - url: 'https://my-url.com'\n" + "    credentialId: myCreds"
            }
            if (!appUrl.url) {
                error "[${STEP_NAME}] No url property was defined for the following element in appUrls: ${appUrl}"
            }
            if (appUrl.credentialId) {
                credentials.add(usernamePassword(credentialsId: appUrl.credentialId, passwordVariable: 'e2e_password', usernameVariable: 'e2e_username', resolveCredentialsId: true))
                if (config.wdi5) {
                    credentials.add(usernamePassword(credentialsId: appUrl.credentialId, passwordVariable: 'wdi5_password', usernameVariable: 'wdi5_username', resolveCredentialsId: true))
                }
            }
        }
    } else{
        if (config.credentialsId) {
            credentials.add(usernamePassword(credentialsId: config.credentialsId, passwordVariable: 'e2e_password', usernameVariable: 'e2e_username', resolveCredentialsId: true))
            if (config.wdi5) {
                credentials.add(usernamePassword(credentialsId: config.credentialsId, passwordVariable: 'wdi5_password', usernameVariable: 'wdi5_username', resolveCredentialsId: true))
            }
        }
    }
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}

