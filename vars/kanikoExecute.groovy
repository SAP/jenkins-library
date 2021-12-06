import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/kanikoExecute.yaml'

void call(Map parameters = [:]) {
    List credentials = [[type: 'file', id: 'dockerConfigJsonCredentialsId', env: ['PIPER_dockerConfigJSON']]]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
