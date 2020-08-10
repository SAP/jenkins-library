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
     * The path to the directory where the credentials file has to be placed.
     */
    'credentialsDirectory'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Writes credentials to a temporary file and deletes it after the body has been executed.
 */
@GenerateDocumentation
void call(Map parameters = [:], Closure body) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        def stageName = parameters.stageName ?: env.STAGE_NAME

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixin(ConfigurationLoader.defaultStageConfiguration(script, stageName))
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        // telemetry reporting
        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], config)

        if (config.credentials && !(config.credentials instanceof List)) {
            error "[${STEP_NAME}] The execution failed, since credentials is not a list. Please provide credentials as a list of maps. For example:\n" +
                "credentials: \n" + "  - alias: 'ERP'\n" + "    credentialId: 'erp-credentials'"
        }
        if (!config.credentialsDirectory) {
            error "[${STEP_NAME}] The execution failed, since no credentialsDirectory is defined. Please provide the path for the credentials file.\n"
        }

        TemporaryCredentialsUtils credUtils = new TemporaryCredentialsUtils(script)

        credUtils.handleTemporaryCredentials(config.credentials, config.credentialsDirectory) {
            body()
        }
    }
}
