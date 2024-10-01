import com.sap.piper.ConfigurationHelper
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/npmExecuteTests.yaml'

@Field Set GLOBAL_CONFIG_KEYS = []

@Field Set STEP_CONFIG_KEYS = [
    "wdi5",
    "credentialsId",
]

@Field Set PARAMETER_KEYS = []

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    String stageName = parameters.stageName ?: env.STAGE_NAME
    Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, PARAMETER_KEYS)
            .mixin(parameters, STEP_CONFIG_KEYS)
            .use()

    List credentials = []
    if (config.credentialsId) {
        if (config.wdi5) {
            credentials.add([type: 'usernamePassword', id: config.credentialsId, env: ['wdi5_username', 'wdi5_password']])
        } else {
            credentials.add([type: 'usernamePassword', id: config.credentialsId, env: ['e2e_username', 'e2e_password']])
        }
    }
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
