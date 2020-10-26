import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/cloudFoundryFaasDeploy.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'usernamePassword', id: 'cfCredentialsId', env: ['PIPER_username', 'PIPER_password']],
        [type: 'token', id: 'xfsrtValuesCredentialsId', env: ['PIPER_xfsrtValues']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials, false, false, true)
}
