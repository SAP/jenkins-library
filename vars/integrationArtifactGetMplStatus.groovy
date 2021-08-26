import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/integrationArtifactGetMplStatus.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'token', id: 'cpiApiServiceKeyCredentialsId', env: ['PIPER_apiServiceKey']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
