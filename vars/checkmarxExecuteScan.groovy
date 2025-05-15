import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/checkmarxExecuteScan.yaml'

//Metadata maintained in file project://resources/metadata/checkmarxExecuteScan.yaml

void call(Map parameters = [:]) {
    List credentials = [[type: 'usernamePassword', id: 'checkmarxCredentialsId', env: ['PIPER_username', 'PIPER_password']], [type: 'token', id: 'githubTokenCredentialsId', env: ['PIPER_githubToken']]]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
