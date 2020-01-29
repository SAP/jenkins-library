import static com.sap.piper.Prerequisites.checkScript
import com.sap.piper.PiperGoUtils
import com.sap.piper.JenkinsUtils
import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import hudson.AbortException
import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/abapEnvironmentPullGitRepo.yaml'
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters, failOnError: true) {

        def script = checkScript(this, parameters) ?: this

        Set configKeys = ['dockerImage', 'dockerWorkspace']
        Map jenkinsConfig = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, configKeys)
            .mixinStepConfig(script.commonPipelineEnvironment, configKeys)
            .mixinStageConfig(script.commonPipelineEnvironment, env.STAGE_NAME, configKeys)
            .mixin(parameters, configKeys)
            .use()

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
            dockerExecute(
                script: script,
                dockerImage: jenkinsConfig.dockerImage,
                dockerWorkspace: jenkinsConfig.dockerWorkspace
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
