import com.sap.piper.Credential
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/kubernetesdeploy.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: Credential.FILE, id: 'kubeConfigFileCredentialsId', env: ['PIPER_kubeConfig']],
        [type: Credential.TOKEN, id: 'kubeTokenCredentialsId', env: ['PIPER_kubeToken']],
        [type: Credential.USERNAME_PASSWORD, id: 'dockerCredentialsId', env: ['PIPER_containerRegistryUser', 'PIPER_containerRegistryPassword']],
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
