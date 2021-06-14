import com.sap.piper.CredentialType
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/kubernetesdeploy.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: CredentialType.FILE, id: 'kubeConfigFileCredentialsId', env: ['PIPER_kubeConfig']],
        [type: CredentialType.FILE, id: 'dockerConfigJsonCredentialsId', env: ['PIPER_dockerConfigJSON']],
        [type: CredentialType.TOKEN, id: 'kubeTokenCredentialsId', env: ['PIPER_kubeToken']],
        [type: CredentialType.USERNAME_PASSWORD, id: 'dockerCredentialsId', env: ['PIPER_containerRegistryUser', 'PIPER_containerRegistryPassword']],
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
