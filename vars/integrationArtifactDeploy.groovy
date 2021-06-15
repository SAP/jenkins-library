import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/integrationArtifactDeploy.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'token', id: 'cpiServiceKeyCredentialId', env: ['PIPER_serviceKey']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
