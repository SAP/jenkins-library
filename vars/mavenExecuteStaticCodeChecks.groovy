import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String METADATA_FILE = 'metadata/mavenStaticCodeChecks.yaml'
@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FOLDER = '.pipeline' // metadata file contains already the "metadata" folder level, hence we end up in a folder ".pipeline/metadata"


void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        final script = checkScript(this, parameters) ?: null

        if (!script) {
            error "Reference to surrounding pipeline script not provided (script: this)."
        }
        def utils = parameters.juStabUtils ?: new Utils()
        new PiperGoUtils(this, utils).unstashPiperBin()

        // Make a shallow copy of the passed-in Map in order to prevent removal of top-level keys
        // to be visible in calling code, just in case the map is still used there.
        parameters = [:] << parameters

        // do not forward these parameters to the go layer
        parameters.remove('juStabUtils')
        parameters.remove('piperGoUtils')
        parameters.remove('script')


        script.commonPipelineEnvironment.writeToDisk(script)
        writeFile(file: "${METADATA_FOLDER}/${METADATA_FILE}", text: libraryResource(METADATA_FILE))

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(parameters)}",
        ]) {
            // get context configuration
            Map contextConfig = readJSON(text: sh(returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '${METADATA_FOLDER}/${METADATA_FILE}'"))

            dockerExecute([script: script].plus([dockerImage: contextConfig.dockerImage])) {
                sh "./piper mavenExecuteStaticCodeChecks"
            }
        }
    }
}
