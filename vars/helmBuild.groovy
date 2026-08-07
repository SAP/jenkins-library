import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = "metadata/helmBuild.yaml"

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'file', id: 'kubeConfigFileCredentialsId', env: ['PIPER_kubeConfig']],
        [type: 'file', id: 'dockerConfigJsonCredentialsId', env: ['PIPER_dockerConfigJSON']],
        [type: 'usernamePassword', id: 'sourceRepositoryCredentialsId', env: ['PIPER_sourceRepositoryUser', 'PIPER_sourceRepositoryPassword']],
        [type: 'usernamePassword', id: 'targetRepositoryCredentialsId', env: ['PIPER_targetRepositoryUser', 'PIPER_targetRepositoryPassword']],
    ]
    // Use helmExecute as the binary step name for compatibility with released binaries
    // predating the helmBuild rename (v1.519.0 and earlier only register helmExecute).
    // The cliAliases in metadata ensure helmExecute routes to helmBuild in new binaries.
    // Revert to STEP_NAME once a binary release with helmBuild ships.
    piperExecuteBin(parameters, 'helmExecute', METADATA_FILE, credentials)
}
