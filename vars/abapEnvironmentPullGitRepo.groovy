import com.sap.piper.MapUtils
import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field def STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/abapEnvironmentPullGitRepo.yaml'

void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters, failOnError: true) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        parameters.juStabUtils = null

        new PiperGoUtils(this, utils).unstashPiperBin()
        utils.unstash('pipelineConfigAndTests')
        script.commonPipelineEnvironment.writeToDisk(script)

        writeFile(file: ".pipeline/tmp/${METADATA_FILE}", text: libraryResource(METADATA_FILE))

        withEnv([
            "PIPER_parametersJSON=${getParametersJSON(parameters)}",
        ]) {
            // get context configuration
            Map config = readJSON (text: sh(returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '.pipeline/tmp/${METADATA_FILE}'"))

            // execute step
            dockerExecute(
                script: script,
                dockerImage: config.dockerImage,
                dockerWorkspace: config.dockerWorkspace
            ) {
                withCredentials([usernamePassword(
                    credentialsId: config.credentialsId,
                    passwordVariable: 'PIPER_password',
                    usernameVariable: 'PIPER_username'
                )]) {
                    sh "./piper abapEnvironmentPullGitRepo"
                }
            }
        }
    }
}

String getParametersJSON(Map parameters = [:]){
    Map stepParameters = [:].plus(parameters)
    // Remove script parameter etc.
    stepParameters.remove('script')
    stepParameters.remove('juStabUtils')
    stepParameters.remove('jenkinsUtilsStub')
    // When converting to JSON and back again, entries which had a 'null' value will now have a value
    // of type 'net.sf.json.JSONNull', for which the Groovy Truth resolves to 'true' in for example if-conditions
    stepParameters = MapUtils.pruneNulls(stepParameters)
    return groovy.json.JsonOutput.toJson(stepParameters)
}
