import com.sap.piper.Credential
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/githubstatus.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: Credential.TOKEN, id: 'githubTokenCredentialsId', env: ['PIPER_token']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
