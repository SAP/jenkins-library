import com.sap.piper.BuildTool
import com.sap.piper.DownloadCacheUtils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript
import static groovy.json.JsonOutput.toJson

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/npmExecuteScripts.yaml'

//Metadata maintained in file project://resources/metadata/npmExecuteScripts.yaml

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this

    // No credentials required/supported as of now
    List credentials = []
    parameters.dockerOptions = ['--cap-add=SYS_ADMIN'].plus(parameters.dockerOptions?:[])
    parameters = DownloadCacheUtils.injectDownloadCacheInParameters(script, parameters, BuildTool.NPM)
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
