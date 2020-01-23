import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/kubernetesdeploy.yaml'

void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this

        def utils = parameters.juStabUtils ?: new Utils()
        parameters.juStabUtils = null

        new PiperGoUtils(this, utils).unstashPiperBin()
        utils.unstash('pipelineConfigAndTests')
        script.commonPipelineEnvironment.writeToDisk(script)

        writeFile(file: METADATA_FILE, text: libraryResource(METADATA_FILE))

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(parameters)}",
        ]) {
            // get context configuration
            Map config = readJSON (text: sh(returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '${METADATA_FILE}'"))
            echo "Config: ${config}"

            dockerExecute(
                script: script,
                dockerImage: config.dockerImage,
                dockerWorkspace: config.dockerWorkspace,
            ) {
                def creds = []
                if (config.kubeConfigFileCredentialsId) creds.add(file(credentialsId: config.kubeConfigFileCredentialsId, variable: 'PIPER_kubeConfig'))
                if (config.kubeTokenCredentialsId) creds.add(string(credentialsId: config.kubeTokenCredentialsId, variable: 'PIPER_kubeToken'))
                if (config.dockerCredentialsId) creds.add(usernamePassword(credentialsId: config.dockerCredentialsId, passwordVariable: 'PIPER_containerRegistryPassword', usernameVariable: 'PIPER_containerRegistryUser'))

                // execute step
                withCredentials(creds) {
                    sh "./piper kubernetesDeploy"
                }
            }
        }
    }
}
