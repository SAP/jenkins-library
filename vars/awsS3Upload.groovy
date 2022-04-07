import groovy.transform.Field
import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/awsS3Upload.yaml'

void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this

    List credentials = [
        [type: 'file', id: 'awsFileCredentialsId', env: ['PIPER_jsonKeyFilePath']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}