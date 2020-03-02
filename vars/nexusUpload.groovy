import static com.sap.piper.Prerequisites.checkScript

import static groovy.json.JsonOutput.toJson

import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils

import groovy.transform.Field

@Field String METADATA_FILE = 'metadata/nexusUpload.yaml'
@Field String STEP_NAME = getClass().getName()


void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        final Script script = checkScript(this, parameters) ?: null
        if (!script) {
            error "Reference to surrounding pipeline script not provided (script: this)."
        }

        def utils = parameters.juStabUtils ?: new Utils()
        parameters.remove('juStabUtils')
        parameters.remove('jenkinsUtilsStub')

        if (!parameters.get('credentialsId')) {
            // Remove null or empty credentialsId key. (Eases calling code.)
            parameters.remove('credentialsId')
        }

        new PiperGoUtils(this, utils).unstashPiperBin()
        utils.unstash('pipelineConfigAndTests')
        script.commonPipelineEnvironment.writeToDisk(script)

        writeFile(file: METADATA_FILE, text: libraryResource(METADATA_FILE))

        // Replace 'artifacts' List with JSON encoded String
        if (parameters.artifacts) {
            parameters.artifacts = "${toJson(parameters.artifacts as List)}"
        }
        // Replace 'additionalClassifiers' List with JSON encoded String
        if (parameters.additionalClassifiers) {
            parameters.additionalClassifiers = "${toJson(parameters.additionalClassifiers as List)}"
        }
        // TODO: This should be handled in the Piper nexusUpload cmd implementation instead!
        // But from the code of commonPipelineEnvironment.writeToDisk() it isn't clear to me whether this would be persisted...
        if (!parameters.artifactId && script.commonPipelineEnvironment.configuration.artifactId) {
            parameters.artifactId = script.commonPipelineEnvironment.configuration.artifactId
        }

        parameters.remove('script')

        withEnv([
            "PIPER_parametersJSON=${toJson(parameters)}",
        ]) {
            // get context configuration
            Map config = readJSON (text: sh(returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '${METADATA_FILE}'"))

            // execute step
            if (config.credentialsId) {
                withCredentials([usernamePassword(
                    credentialsId: config.credentialsId,
                    passwordVariable: 'PIPER_password',
                    usernameVariable: 'PIPER_username'
                )]) {
                    sh "./piper nexusUpload"
                }
            } else {
                sh "./piper nexusUpload"
            }
        }
    }
}
