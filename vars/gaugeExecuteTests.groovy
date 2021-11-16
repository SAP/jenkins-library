import static com.sap.piper.Prerequisites.checkScript
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/gaugeExecuteTests.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'usernamePassword', id: 'seleniumHubCredentialsId', env: ['PIPER_SELENIUM_HUB_USER', 'PIPER_SELENIUM_HUB_PASSWORD']],
    ]
    final script = checkScript(this, parameters) ?: this
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
