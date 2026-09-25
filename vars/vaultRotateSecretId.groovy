import groovy.transform.Field
import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/vaultRotateSecretId.yaml'

void call(Map parameters = [:]) {
        def script = checkScript(this, parameters) ?: this
        List credentials = [
            [type: 'token', id: 'jenkinsUrlCredentialsId', env: ['PIPER_jenkinsUrl']],
            [type: 'token', id: 'jenkinsUsernameCredentialsId', env: ['PIPER_jenkinsUsername']],
            [type: 'token', id: 'jenkinsTokenCredentialsId', env: ['PIPER_jenkinsToken']]
        ]
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials, false, false, false)
}
