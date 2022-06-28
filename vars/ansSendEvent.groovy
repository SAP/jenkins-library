import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/ansSendEvent.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'token', id: 'ansServiceKeyCredentialsId', env: ['PIPER_ansServiceKey']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
