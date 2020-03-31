import com.sap.piper.DownloadCacheUtils
import groovy.transform.Field

import static groovy.json.JsonOutput.toJson

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/nexusUpload.yaml'

//Metadata maintained in file project://resources/metadata/nexusUpload.yaml

void call(Map parameters = [:]) {
    // Replace 'additionalClassifiers' List with JSON encoded String.
    // This is currently necessary, since the go code doesn't support complex/arbitrary parameter types.
    // TODO: Support complex/structured types of parameters in piper-go
    if (parameters.additionalClassifiers) {
        parameters.additionalClassifiers = "${toJson(parameters.additionalClassifiers as List)}"
    }
    parameters = DownloadCacheUtils.injectDownloadCacheInMavenParameters(parameters.script, parameters)

    List credentials = [[type: 'usernamePassword', id: 'nexusCredentialsId', env: ['PIPER_username', 'PIPER_password']]]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
