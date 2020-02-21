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

        echo "nexusUpload parameters: $parameters"

        sh 'git clone https://github.com/SAP/jenkins-library.git -b nexus-upload .piper-git-checkout'
        dir('.piper-git-checkout') {
            dockerExecute(script: this, dockerImage: 'golang:1.13', dockerOptions: '-u 0') {
                sh 'go build -o piper . && chmod +x piper && mv piper ..'
            }
        }
        sh 'rm -rf .piper-git-checkout'

//        new PiperGoUtils(this, utils).unstashPiperBin()
//        utils.unstash('pipelineConfigAndTests')
        script.commonPipelineEnvironment.writeToDisk(script)

        writeFile(file: METADATA_FILE, text: libraryResource(METADATA_FILE))

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(parameters)}",
        ]) {
            //sh 'env'
            // get context configuration
            Map config = readJSON (text: sh(returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '${METADATA_FILE}'"))

            // Hack to get things going (reading config from ENV doesn't work for some reason):
            // config = parameters

            Closure body = {
                // String url = config.url
                // String repository = config.repository
                // String version = config.version
                // String groupId = config.groupId

                // // config.artifacts is supposed to be a List of Map objects, where each Map contains the
                // // keys 'artifactId', 'classifier', 'type' and 'file'.
                // String artifacts = toJson(config.artifacts as List)
                // artifacts = artifacts.replace('"', '\\"')

//                sh "./piper nexusUpload --url=$url --repository=$repository --groupId=$groupId --version=$version --artifacts=\"$artifacts\""
                sh "./piper nexusUpload"
            }

            // execute step
            if (config.credentialsId) {
                withCredentials([usernamePassword(
                    credentialsId: config.credentialsId,
                    passwordVariable: 'PIPER_password',
                    usernameVariable: 'PIPER_username'
                )]) {
                    body.call()
                }
            } else {
                body.call()
            }
        }
    }
}
