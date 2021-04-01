import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/terraformExecute.yaml'

void call(Map parameters = [:]) {
    List credentials = [[type: 'file', id: 'terraformSecrets', env: ['PIPER_terraformSecrets']]]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
