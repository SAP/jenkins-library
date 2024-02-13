import groovy.transform.Field
import com.sap.piper.JenkinsUtils

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/tmsExport.yaml'

void call(Map parameters = [:]) {
    List credentials = [
        [type: 'token', id: 'credentialsId', env: ['PIPER_serviceKey']]
    ]

    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials, false, false, true)
}
