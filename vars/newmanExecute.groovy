import com.sap.piper.ConfigurationHelper
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/newmanExecute.yaml'

@Field Set CONFIG_KEYS = [
    /**
     * Define name array of cloud foundry apps deployed for which secrets (clientid and clientsecret) will be appended
     * to the newman command that overrides the environment json entries
     * (--env-var <appName_clientid>=${clientid} & --env-var <appName_clientsecret>=${clientsecret})
     */
    "cfAppsWithSecrets",
]

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    String stageName = parameters.stageName ?: env.STAGE_NAME
    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults([:], stageName)
        .mixinGeneralConfig(script.commonPipelineEnvironment, CONFIG_KEYS)
        .mixinStepConfig(script.commonPipelineEnvironment, CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, CONFIG_KEYS)
        .use()

    List credentials = []
    if (config.cfAppsWithSecrets) {
        config.cfAppsWithSecrets.each {
            echo "[INFO]${STEP_NAME}] Preparing credential for being used by piper-go. key: ${it}, exposed as environment variable PIPER_NEWMAN_USER_${it} and PIPER_NEWMAN_PASSWORD_${it}"
            credentials << [type: 'usernamePassword', id: "${it}", env: ["PIPER_NEWMAN_USER_${it}", "PIPER_NEWMAN_PASSWORD_${it}"]]
        }
    }
    print credentials
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
