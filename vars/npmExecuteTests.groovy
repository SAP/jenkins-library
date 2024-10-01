import com.sap.piper.ConfigurationHelper
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/npmExecuteTests.yaml'

@Field Set CONFIG_KEYS = [
    "wdi5",
    "credentialsId",
]

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    String stageName = parameters.stageName ?: env.STAGE_NAME
    Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, CONFIG_KEYS)
            .mixin(parameters, CONFIG_KEYS)
            .use()

    List credentials = []
    if config.credentialsId {
        if config.wdi5 {
            credentials.add([type: 'usernamePassword', id: config.credentialsId, env: ['wdi5_username', 'wdi5_password']])
        } else {
            credentials.add([type: 'usernamePassword', id: config.credentialsId, env: ['e2e_username', 'e2e_password']]
        }
    }
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
