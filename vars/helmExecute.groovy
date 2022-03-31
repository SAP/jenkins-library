import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = "metadata/helmExecute.yaml"

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'file', id: 'kubeConfigFileCredentialsId', env: ['PIPER_kubeConfig']],
        [type: 'file', id: 'dockerConfigJsonCredentialsId', env: ['PIPER_dockerConfigJSON']],
        [type: 'usernamePassword', id: 'dockerCredentialsId', env: ['PIPER_containerRegistryUser', 'PIPER_containerRegistryPassword']],
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
