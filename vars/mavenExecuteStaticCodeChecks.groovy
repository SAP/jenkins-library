import com.sap.piper.Utils
import groovy.transform.Field

@Field String GO_COMMAND = 'mavenStaticCodeChecks'
@Field String METADATA_FILE = 'metadata/mavenStaticCodeChecks.yaml'
@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FOLDER = '.pipeline' // metadata file contains already the "metadata" folder level, hence we end up in a folder ".pipeline/metadata"

void call(Map parameters = [:]) {
    def utils = parameters.juStabUtils ?: new Utils()
    utils.runPiperGoStep(this, parameters)
}
