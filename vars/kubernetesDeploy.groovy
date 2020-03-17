import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/kubernetesdeploy.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'file', id: 'kubeConfigFileCredentialsId', env: ['PIPER_kubeConfig']],
        [type: 'token', id: 'kubeTokenCredentialsId', env: ['PIPER_kubeToken']],
        [type: 'usernamePassword', id: 'dockerCredentialsId', env: ['PIPER_containerRegistryUser', 'PIPER_containerRegistryPassword']],
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
