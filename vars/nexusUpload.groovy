import static com.sap.piper.Prerequisites.checkScript

import static groovy.json.JsonOutput.toJson

import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils

import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/nexusUpload.yaml'

//Metadata maintained in file project://resources/metadata/nexusUpload.yaml

void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        final Script script = checkScript(this, parameters) ?: null
        if (!script) {
            error "Reference to surrounding pipeline script not provided (script: this)."
        }

        def utils = parameters.juStabUtils ?: new Utils()

        // Make shallow copy of parameters, so we can add/remove (top-level) keys without side-effects to calling code
        parameters = [:] << parameters
        parameters.remove('juStabUtils')
        parameters.remove('jenkinsUtilsStub')

        // Backwards compatibility
        if (parameters.credentialsId && !parameters.nexusCredentialsId) {
            parameters.nexusCredentialsId = parameters.credentialsId
        }
        parameters.remove('credentialsId')
        // Remove empty credentials, since the will end up as "net.sf.json.JSONNull"
        // when reading back the config via "piper getConfig --contextConfig" and
        // that in turn will trigger the withCredentials() code-path, but fail to
        // create a binding.
        if (!parameters.nexusCredentialsId) {
            parameters.remove('nexusCredentialsId')
        }

        if (!fileExists('./piper')) {
            new PiperGoUtils(this, utils).unstashPiperBin()
        }
        utils.unstash('pipelineConfigAndTests')
        script.commonPipelineEnvironment.writeToDisk(script)

        writeFile(file: METADATA_FILE, text: libraryResource(METADATA_FILE))

        // Replace 'additionalClassifiers' List with JSON encoded String
        if (parameters.additionalClassifiers) {
            parameters.additionalClassifiers = "${toJson(parameters.additionalClassifiers as List)}"
        }
        // Fall-back to artifactId from configuration if not given
        if (!parameters.artifactId && script.commonPipelineEnvironment.configuration.artifactId) {
            parameters.artifactId = script.commonPipelineEnvironment.configuration.artifactId
        }
        // Pass MTAR file path
        if (!parameters.artifactId && script.commonPipelineEnvironment.configuration.artifactId) {
            parameters.artifactId = script.commonPipelineEnvironment.configuration.artifactId
        }

        parameters.remove('script')

        withEnv([
            "PIPER_parametersJSON=${toJson(parameters)}",
        ]) {
            // get context configuration
            Map config = readJSON (text: sh(returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '${METADATA_FILE}'"))

            sh 'ls -laR .pipeline'

            // execute step
            if (config.nexusCredentialsId) {
                withCredentials([usernamePassword(
                    credentialsId: config.nexusCredentialsId,
                    passwordVariable: 'PIPER_password',
                    usernameVariable: 'PIPER_username'
                )]) {
                    sh "./piper nexusUpload --verbose"
                }
            } else {
                sh "./piper nexusUpload --verbose"
            }
        }
    }
}
