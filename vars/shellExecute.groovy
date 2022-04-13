import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/shellExecute.yaml'

void call(Map parameters = [:]) {
     List credentials = [[type: 'token', id: 'githubTokenCredentialsId', env: ['PIPER_githubToken']]]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
