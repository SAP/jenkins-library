import static com.sap.piper.Prerequisites.checkScript
import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.JenkinsUtils
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
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()
        String piperGoPath = parameters.piperGoPath ?: './piper'

        piperExecuteBin.prepareExecution(script, utils, parameters)
        piperExecuteBin.prepareMetadataResource(script, METADATA_FILE)
        Map stepParameters = piperExecuteBin.prepareStepParameters(parameters)

        List credentialInfo = [
            // [type: 'token', id: 'githubTokenCredentialsId', env: ['PIPER_githubToken']],
        ]

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(stepParameters)}",
            "PIPER_correlationID=${env.BUILD_URL}",
        ]) {
            String customDefaultConfig = piperExecuteBin.getCustomDefaultConfigsArg()
            String customConfigArg = piperExecuteBin.getCustomConfigArg(script)
            // get context configuration
            Map config
            piperExecuteBin.handleErrorDetails(STEP_NAME) {
                config = piperExecuteBin.getStepContextConfig(script, piperGoPath, METADATA_FILE, customDefaultConfig, customConfigArg)
                echo "Context Config: ${config}"
            }

            piperExecuteBin.dockerWrapper(script, STEP_NAME, config){
                piperExecuteBin.handleErrorDetails(STEP_NAME) {
                    script.commonPipelineEnvironment.writeToDisk(script)
                    try {
                        piperExecuteBin.credentialWrapper(config, credentialInfo){
                            // sh "${piperGoPath} ${STEP_NAME}${customDefaultConfig}${customConfigArg}"
                            sh "${piperGoPath} hadolintExecuteScan${customDefaultConfig}${customConfigArg}"
                        }
                    } finally {
                        jenkinsUtils.handleStepResults(STEP_NAME, true, false)
                        script.commonPipelineEnvironment.readFromDisk(script)

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
    }
}
