import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/newmanExecute.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        //[type: 'usernamePassword', id: 'seleniumHubCredentialsId', env: ['PIPER_SELENIUM_HUB_USER', 'PIPER_SELENIUM_HUB_PASSWORD']],
    ]

    if (config.cfAppsWithSecrets) {
        config.cfAppsWithSecrets.each {
            echo "[INFO]${STEP_NAME}] Preparing credential for being used by piper-go. key: ${it}, exposed as environment variable PIPER_NEWMAN_USER_${it} and PIPER_NEWMAN_PASSWORD_${it}"
            credentials << [type: 'usernamePassword', id: ${it}, env: ["PIPER_NEWMAN_USER_${it}", "PIPER_NEWMAN_PASSWORD_${it}"]]
        }
    }

    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
