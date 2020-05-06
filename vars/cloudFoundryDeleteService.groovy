import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/cloudFoundryDeleteService.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'usernamePassword', id: 'cfCredentialsId', env: ['PIPER_username', 'PIPER_password']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials, false, false, true)
}
