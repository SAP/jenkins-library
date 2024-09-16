import com.sap.piper.ConfigurationHelper
import com.sap.piper.DownloadCacheUtils
import com.sap.piper.GenerateDocumentation
import com.sap.piper.k8s.ContainerMap
import groovy.transform.Field
import com.sap.piper.Utils

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/npmExecuteEndToEndTests.yaml'

void call(Map parameters = [:]) {
    if (parameters.appUrls && !(parameters.appUrls instanceof List)) {
        error "[${STEP_NAME}] The execution failed, since appUrls is not a list. Please provide appUrls as a list of maps. For example:\n" +
                "appUrls: \n" + "  - url: 'https://my-url.com'\n" + "    credentialId: myCreds"
    }
    List credentials = []
    if (parameters.appUrls){
        for (int i = 0; i < parameters.appUrls.size(); i++) {
            def appUrl = parameters.appUrls[i]
            if (!(appUrl instanceof Map)) {
                error "[${STEP_NAME}] The element ${appUrl} is not of type map. Please provide appUrls as a list of maps. For example:\n" +
                        "appUrls: \n" + "  - url: 'https://my-url.com'\n" + "    credentialId: myCreds"
            }
            if (!appUrl.url) {
                error "[${STEP_NAME}] No url property was defined for the following element in appUrls: ${appUrl}"
            }
            if (appUrl.credentialId) {
                credentials.add(usernamePassword(credentialsId: appUrl.credentialId, passwordVariable: 'e2e_password', usernameVariable: 'e2e_username'))
                if (parameters.wdi5) {
                    credentials.add(usernamePassword(credentialsId: appUrl.credentialId, passwordVariable: 'wdi5_password', usernameVariable: 'wdi5_username'))
                }
            }
        }
    } else{
        if (parameters.credentialsId) {
            credentials.add(usernamePassword(credentialsId: parameters.credentialsId, passwordVariable: 'e2e_password', usernameVariable: 'e2e_username'))
            if (parameters.wdi5) {
                credentials.add(usernamePassword(credentialsId: parameters.credentialsId, passwordVariable: 'wdi5_password', usernameVariable: 'wdi5_username'))
            }
        }
    }
    final script = checkScript(this, parameters) ?: this
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}

