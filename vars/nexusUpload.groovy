import static com.sap.piper.Prerequisites.checkScript

import static groovy.json.JsonOutput.toJson

import com.sap.piper.JenkinsUtils
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
        parameters.juStabUtils = null
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()
        parameters.jenkinsUtilsStub = null

        new PiperGoUtils(this, utils).unstashPiperBin()
        utils.unstash('pipelineConfigAndTests')
        script.commonPipelineEnvironment.writeToDisk(script)

        writeFile(file: METADATA_FILE, text: libraryResource(METADATA_FILE))

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(parameters)}",
        ]) {
            // get context configuration
            Map config = readJSON (text: sh(returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '${METADATA_FILE}'"))

            // execute step
            withCredentials([usernamePassword(
                credentialsId: config.nexusCredentialsId,
                passwordVariable: 'PIPER_password',
                usernameVariable: 'PIPER_username'
            )]) {
                String url = config.url
                String repository = config.repository
                String version = config.version
                String groupId = config.groupId

                // config.artifacts is supposed to be a List of Map objects, where each Map contains the
                // keys 'artifactId', 'classifier', 'type' and 'file'.
                String artifacts = toJson(config.artifacts as List)
                artifacts = artifacts.replace('"', '\\"')
//                artifacts = artifacts.replace(':', '\\:')

                sh "./piper nexusUpload --url=$url --repository=$repository --groupId=$groupId --version=$version --artifacts=\"$artifacts\""
            }

            jenkinsUtils.handleStepResults(STEP_NAME, true, false)
        }
    }
}
