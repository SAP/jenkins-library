import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = "metadata/golangBuild.yaml"

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'usernamePassword', id: 'golangPrivateModulesGitTokenCredentialsId', env: ['PIPER_privateModulesGitUsername', 'PIPER_privateModulesGitToken']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
