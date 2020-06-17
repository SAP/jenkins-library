import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/detect.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'token', id: 'detectTokenCredentialsId', env: ['PIPER_apiToken']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
