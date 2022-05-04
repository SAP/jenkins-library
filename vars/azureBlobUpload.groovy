import groovy.transform.Field
import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/azureBlobUpload.yaml'

void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this

    List credentials = [
        [type: 'token', id: 'azureCredentialsId', env: ['PIPER_jsonCredentialsAzure']]
    ]

    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
