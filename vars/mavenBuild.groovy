import com.sap.piper.Utils
import groovy.transform.Field

@Field String METADATA_FILE = 'metadata/mavenBuild.yaml'
@Field String STEP_NAME = getClass().getName()

void call(Map parameters = [:]) {
    List credentials = [ ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
