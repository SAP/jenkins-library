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
        
        Map config
        def utils = parameters.juStabUtils ?: new Utils()
        parameters.juStabUtils = null

        // telemetry reporting
        utils.pushToSWA([step: STEP_NAME], config)

        new PiperGoUtils(this, utils).unstashPiperBin()
        utils.unstash('pipelineConfigAndTests')

        writeFile(file: METADATA_FILE, text: libraryResource(METADATA_FILE))

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(parameters)}",
            "PIPER_owner=${script.commonPipelineEnvironment.getGithubOrg()?:''}",
            "PIPER_repository=${script.commonPipelineEnvironment.getGithubRepo()?:''}",
            "PIPER_version=${script.commonPipelineEnvironment.getArtifactVersion()?:''}"
        ]) {
            // get context configuration
            config = readJSON (text: sh(returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '${METADATA_FILE}' --stepName ${STEP_NAME}"))

            // execute step
            withCredentials([string(credentialsId: config.githubTokenCredentialsId, variable: 'TOKEN')]) {
                sh "./piper githubPublishRelease  --token ${TOKEN}"
            }
        }
    }
}
