import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/abapEnvironmentCloneGitRepo.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'usernamePassword', id: 'abapCredentialsId', env: ['PIPER_username', 'PIPER_password']],
        [type: 'usernamePassword', id: 'byogCredentialsId', env: ['PIPER_byogUsername', 'PIPER_byogPassword']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials, false, false, true)
}
