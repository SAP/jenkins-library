import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.GitUtils
import com.sap.piper.Utils

import groovy.text.GStringTemplateEngine
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/karmaExecuteTests.yaml'

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = []
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS + []

/**
 * Executes the Karma tests.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    List credentials = [
        [type: 'usernamePassword', id: 'seleniumHubCredentialsId', env: ['PIPER_SELENIUM_HUB_USER', 'PIPER_SELENIUM_HUB_PASSWORD']],
    ]
    final script = checkScript(this, parameters) ?: this
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
