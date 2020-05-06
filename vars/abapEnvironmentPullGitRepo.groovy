import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/abapEnvironmentPullGitRepo.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'usernamePassword', id: 'credentialsId', env: ['PIPER_username', 'PIPER_password']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials, false, false, true)
}
