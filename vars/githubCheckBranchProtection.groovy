import com.sap.piper.CredentialType
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/githubbranchprotection.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: CredentialType.TOKEN, id: 'githubTokenCredentialsId', env: ['PIPER_token']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
