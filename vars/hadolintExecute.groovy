import static com.sap.piper.Prerequisites.checkScript
import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/hadolint.yaml'

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Dockerfile to be used for the assessment.
     */
    'dockerFile',
    /**
     * Name of the docker image that should be used, in which node should be installed and configured. Default value is 'hadolint/hadolint:latest-debian'.
     */
    'dockerImage'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Name of the configuration file used locally within the step. If a file with this name is detected as part of your repo downloading the central configuration via `configurationUrl` will be skipped. If you change the file's name make sure your stashing configuration also reflects this.
     */
    'configurationFile',
    /**
     * URL pointing to the .hadolint.yaml exclude configuration to be used for linting. Also have a look at `configurationFile` which could avoid central configuration download in case the file is part of your repository.
     */
    'configurationUrl',
    /**
     * If the url provided as configurationUrl is protected, this Jenkins credential can be used to authenticate the request.
     */
    'configurationCredentialsId',
    /**
     * Docker options to be set when starting the container.
     */
    'dockerOptions',
    /**
     * Quality Gates to fail the build, see [warnings-ng plugin documentation](https://github.com/jenkinsci/warnings-plugin/blob/master/doc/Documentation.md#quality-gate-configuration).
     */
    'qualityGates',
    /**
     * Name of the result file used locally within the step.
     */
    'reportFile',
    /**
     * Name of the checkstyle report being generated our of the results.
     */
    'reportName'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
/**
 * Executes the Haskell Dockerfile Linter which is a smarter Dockerfile linter that helps you build [best practice](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/) Docker images.
 * The linter is parsing the Dockerfile into an abstract syntax tree (AST) and performs rules on top of the AST.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        final script = checkScript(this, parameters) ?: this
        final utils = parameters.juStabUtils ?: new Utils()
        String stageName = parameters.stageName ?: env.STAGE_NAME

        // load default & individual configuration
        Map config

        new PiperGoUtils(this, utils).unstashPiperBin()
        utils.unstash('pipelineConfigAndTests')
        script.commonPipelineEnvironment.writeToDisk(script)

        writeFile(file: METADATA_FILE, text: libraryResource(METADATA_FILE))

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(parameters)}",
        ]) {
            // get context configuration
            config = readContextConfig()
            // ---- move to go
            if(!fileExists(configuration.configurationFile) && configuration.configurationUrl) {
                downloadFile(configuration.configurationUrl, configuration.configurationFile, configuration.configurationCredentialsId)
                if(existingStashes) {
                    def stashName = 'hadolintConfiguration'
                    stash name: stashName, includes: configuration.configurationFile
                    existingStashes += stashName
                }
            }
            // ----
            // telemetry reporting
            utils.pushToSWA([
                step: STEP_NAME,
                stepParamKey1: 'scriptMissing',
                stepParam1: parameters?.script == null
            ], config)

            utils.unstashAll(config.stashContent)

            dockerExecute(
                script: script,
                dockerImage: config.dockerImage,
                dockerOptions: config.dockerOptions
            ) {
                sh "./piper hadolintExecute"

                jenkinsUtils.handleStepResults(STEP_NAME, true, false)
                recordIssues(
                    tools: [checkStyle(
                        name: config.reportName,
                        pattern: config.reportFile,
                        id: config.reportName
                    )],
                    qualityGates: config.qualityGates,
                    enabledForFailure: true,
                    blameDisabled: true
                )
            }
        }
    }
}

String readContextConfig(){
    def configContent = sh returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '${METADATA_FILE}'")
    return readJSON text: configContent
}

void downloadFile(url, target, authentication = null){
        def response = httpRequest url: url, authentication: authentication, timeout: 20
        writeFile text: response.content, file: target
}
