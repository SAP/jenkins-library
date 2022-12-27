import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/integrationArtifactTransport.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'token', id: 'casApiServiceKeyCredentialsId', env: ['PIPER_casServiceKey']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
