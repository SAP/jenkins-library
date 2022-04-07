import com.sap.piper.BuildTool
import groovy.transform.Field
import static com.sap.piper.Prerequisites.checkScript
import com.sap.piper.DownloadCacheUtils

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/awsS3Upload.yaml'

void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    Map stepParams = PiperGoUtils.prepare(this, script).plus(parameters)

    List credentials = [
        [type: 'file', id: 'awsFileCredentialsId', env: ['PIPER_jsonKeyFilePath']]
    ]
    piperExecuteBin(stepParams, STEP_NAME, METADATA_FILE, credentials)
}