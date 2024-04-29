import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/contrastExecuteScan.yaml'

void call(Map parameters = [:]) {
    List credentials = [
    [type: 'usernamePassword', id: 'userCredentialsId', env: ['PIPER_username', 'PIPER_serviceKey']],
    [type: 'token', id: 'apiKeyCredentialsId', env: ['PIPER_userApiKey']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
