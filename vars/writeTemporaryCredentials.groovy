import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationLoader
import com.sap.piper.GenerateDocumentation
import com.sap.piper.JsonUtils
import com.sap.piper.TemporaryCredentialsUtils
import com.sap.piper.Utils

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = [
    /**
     * The list of credentials that are written to a temporary file for the execution of the body.
     * Each element of credentials must be a map containing a property alias and a property credentialId.
     * You have to ensure that corresponding credential entries exist in your Jenkins configuration.
     */
    'credentials',
    /**
     * The list of paths to directories where credentials files need to be placed.
     */
    'credentialsDirectories'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Writes credentials to a temporary file and deletes it after the body has been executed.
 */
@GenerateDocumentation
void call(Map parameters = [:], Closure body) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        String stageName = parameters.stageName ?: env.STAGE_NAME

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixin(ConfigurationLoader.defaultStageConfiguration(script, stageName))
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        if (config.credentials && !(config.credentials instanceof List)) {
            error "[${STEP_NAME}] The execution failed, since credentials is not a list. Please provide credentials as a list of maps. For example:\n" +
                "credentials: \n" + "  - alias: 'ERP'\n" + "    credentialId: 'erp-credentials'"
        }
        if (!config.credentialsDirectories) {
            error "[${STEP_NAME}] The execution failed, since no credentialsDirectories are defined. Please provide a list of paths for the credentials files.\n"
        }
        if (!(config.credentialsDirectories  instanceof List)) {
            error "[${STEP_NAME}] The execution failed, since credentialsDirectories is not a list. Please provide credentialsDirectories as a list of paths.\n"
        }

        TemporaryCredentialsUtils credUtils = new TemporaryCredentialsUtils(script)

        credUtils.handleTemporaryCredentials(config.credentials, config.credentialsDirectories) {
            body()
        }
    }
}
