import com.sap.piper.DownloadCacheUtils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript
import static groovy.json.JsonOutput.toJson

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/jsBuild.yaml'

//Metadata maintained in file project://resources/metadata/jsBuild.yaml

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this

    // No credentials required/supported as of now
    List credentials = []

    parameters['dockerOptions'] = DownloadCacheUtils.getDockerOptions(script)
    parameters['defaultNpmRegistry'] = "http://${script.env.DL_CACHE_HOSTNAME}:8081/repository/npm-proxy/"
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
