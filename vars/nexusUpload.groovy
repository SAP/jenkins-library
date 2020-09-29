import com.sap.piper.BuildTool
import com.sap.piper.DownloadCacheUtils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/nexusUpload.yaml'

//Metadata maintained in file project://resources/metadata/nexusUpload.yaml

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    parameters = DownloadCacheUtils.injectDownloadCacheInParameters(script, parameters, BuildTool.MAVEN)

    List credentials = [[type: 'usernamePassword', id: 'nexusCredentialsId', env: ['PIPER_username', 'PIPER_password']]]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
