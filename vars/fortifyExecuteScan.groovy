import com.sap.piper.DownloadCacheUtils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/fortify.yaml'

//Metadata maintained in file project://resources/metadata/fortify.yaml

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    parameters = DownloadCacheUtils.injectDownloadCacheInMavenParameters(script, parameters)

    List credentials = [[type: 'token', id: 'fortifyCredentialsId', env: ['PIPER_authToken']]]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
