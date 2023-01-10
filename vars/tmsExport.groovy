import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/tmsExport.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'token', id: 'credentialsId', env: ['PIPER_tmsServiceKey']]
    ]

    def namedUser = jenkinsUtils.getJobStartedByUserId()
    if (namedUser) {
        parameters.namedUser = namedUser
    }
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials, false, false, true)
}