import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/newmanExecute.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        //[type: 'usernamePassword', id: 'seleniumHubCredentialsId', env: ['PIPER_SELENIUM_HUB_USER', 'PIPER_SELENIUM_HUB_PASSWORD']],
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
