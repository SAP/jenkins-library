import com.sap.piper.BuildTool
import com.sap.piper.DownloadCacheUtils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/versioning.yaml'

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this

    List credentials = [
        [type: 'ssh', id: 'gitSshKeyCredentialsId'],
        [type: 'usernamePassword', id: 'gitHttpsCredentialsId', env: ['PIPER_username', 'PIPER_password']],
    ]
    parameters = DownloadCacheUtils.injectDownloadCacheInParameters(script, parameters, BuildTool.MAVEN)
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials, false, false, true)
}
