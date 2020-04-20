import com.sap.piper.DownloadCacheUtils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript
import static groovy.json.JsonOutput.toJson

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/jsBuild.yaml'

//Metadata maintained in file project://resources/metadata/jsBuild.yaml

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
   //todo set env vars for npm parameters = DownloadCacheUtils.injectDownloadCacheInMavenParameters(script, parameters)

    List credentials = []
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
