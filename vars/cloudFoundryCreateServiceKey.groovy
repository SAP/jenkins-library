import com.sap.piper.GenerateDocumentation
import com.sap.piper.BashUtils
import com.sap.piper.PiperGoUtils
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field def STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/cloudFoundryCreateServiceKey.yaml'

void call(Map parameters = [:]) {
        List credentials = [
        [type: 'usernamePassword', id: 'cfCredentialsId', env: ['PIPER_username', 'PIPER_password']],
        ]
        
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
