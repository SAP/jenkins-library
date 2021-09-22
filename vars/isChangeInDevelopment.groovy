import com.sap.piper.BuildTool
import groovy.transform.Field
import static com.sap.piper.Prerequisites.checkScript
import com.sap.piper.DownloadCacheUtils

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/isChangeInDevelopment.yaml'

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    parameters = DownloadCacheUtils.injectDownloadCacheInParameters(script, parameters, BuildTool.MTA)

    List credentials = [
        [type: 'usernamePassword', id: 'credentialsId', env: ['PIPER_username', 'PIPER_password']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
