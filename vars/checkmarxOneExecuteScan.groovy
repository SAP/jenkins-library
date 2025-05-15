import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/checkmarxOneExecuteScan.yaml'

//Metadata maintained in file project://resources/metadata/checkmarxoneExecuteScan.yaml

void call(Map parameters = [:]) {
    List credentials = [[type: 'usernamePassword', id: 'checkmarxOneCredentialsId', env: ['PIPER_clientId', 'PIPER_clientSecret']],
                        [type: 'token', id: 'checkmarxOneAPIKey', env: ['PIPER_APIKey']]]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
