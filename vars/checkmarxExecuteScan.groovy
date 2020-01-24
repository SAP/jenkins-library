import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/checkmarx.yaml'

//Metadata maintained in file project://resources/metadata/checkmarx.yaml

void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this

        Map config
        def utils = parameters.juStabUtils ?: new Utils()
        parameters.juStabUtils = null

        // telemetry reporting
        utils.pushToSWA([step: STEP_NAME], config)

        new PiperGoUtils(this, utils).unstashPiperBin()
        utils.unstash('pipelineConfigAndTests')
        script.commonPipelineEnvironment.writeToDisk(script)

        writeFile(file: METADATA_FILE, text: libraryResource(METADATA_FILE))

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(parameters)}",
        ]) {
            // get context configuration
            config = readJSON (text: sh(returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '${METADATA_FILE}'"))

            // execute step
            withCredentials([usernamePassword(
                credentialsId: config.checkmarxCredentialsId,
                passwordVariable: 'PIPER_password',
                usernameVariable: 'PIPER_username'
            )]) {
                sh "./piper checkmarxExecuteScan --verbose ${config.verbose}"
            }

            if (config.generatePdfReport)
                archiveArtifacts artifacts: "**/CxSASTReport_*.pdf", allowEmptyArchive: false

            archiveArtifacts artifacts: "**/CxSASTResults_*.xml", allowEmptyArchive: true
        }
    }
}
