import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/abapEnvironmentRunATCCheck.yaml'

void call(Map parameters = [:]) {
        List credentials = [
        [type: 'usernamePassword', id: 'abapCredentialsId', env: ['PIPER_username', 'PIPER_password']]
        ]
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials, true, false, true)
}
