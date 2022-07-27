import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = ""

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'usernamePassword', id: 'checkIfStepActiveCredentialsId', env: ['PIPER_username', 'PIPER_password']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, [])
}
