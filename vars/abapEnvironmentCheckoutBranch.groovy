import com.sap.piper.CredentialType
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/abapEnvironmentCheckoutBranch.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: CredentialType.USERNAME_PASSWORD, id: 'abapCredentialsId', env: ['PIPER_username', 'PIPER_password']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials, false, false, true)
}
