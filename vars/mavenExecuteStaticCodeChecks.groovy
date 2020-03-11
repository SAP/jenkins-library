import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String METADATA_FILE = 'metadata/mavenExecuteStaticCodeChecks.yaml'
@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FOLDER = '.pipeline' // metadata file contains already the "metadata" folder level, hence we end up in a folder ".pipeline/metadata"


void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        final script = checkScript(this, parameters) ?: null

        if (!script) {
            error "Reference to surrounding pipeline script not provided (script: this)."
        }

        // The parameters map in provided from outside. That map might be used elsewhere in the pipeline
        // hence we should not modify it here. So we create a new map based on the parameters map.
        parameters = [:] << parameters

        // hard to predict how these parameters looks like in its serialized form. Anyhow it is better
        // not to have these parameters forwarded somehow to the go layer.
        parameters.remove('juStabUtils')
        parameters.remove('piperGoUtils')
        parameters.remove('script')

        def utils = parameters.juStabUtils ?: new Utils()
        def piperGoUtils = parameters.piperGoUtils ?: new PiperGoUtils(utils)
        piperGoUtils.unstashPiperBin()

        script.commonPipelineEnvironment.writeToDisk(script)
        writeFile(file: "${METADATA_FOLDER}/${METADATA_FILE}", text: libraryResource(METADATA_FILE))

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(parameters)}",
        ]) {
            // get context configuration
            Map contextConfig = readJSON(text: sh(returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '${METADATA_FOLDER}/${METADATA_FILE}'"))

            dockerExecute([script: script].plus([dockerImage: contextConfig.dockerImage])) {
                sh "./piper MavenExecuteStaticCodeChecks"
            }
        }
    }
}
