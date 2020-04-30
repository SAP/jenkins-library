import com.sap.piper.MapUtils
import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/githubrelease.yaml'

//Metadata maintained in file project://resources/metadata/githubrelease.yaml

void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        parameters.juStabUtils = null

        // telemetry reporting
        utils.pushToSWA([step: STEP_NAME], config)

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
            withCredentials([string(credentialsId: config.githubTokenCredentialsId, variable: 'PIPER_token')]) {
                sh './piper githubPublishRelease'
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
