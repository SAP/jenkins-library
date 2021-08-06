import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/integrationArtifactTriggerIntegrationTest.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'token', id: 'integrationFlowServiceKeyCredentialsId', env: ['PIPER_integrationFlowServiceKey']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
