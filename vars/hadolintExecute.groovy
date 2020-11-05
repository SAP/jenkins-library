import static com.sap.piper.Prerequisites.checkScript
import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/hadolint.yaml'

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
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
            [type: 'usernamePassword', id: 'configurationCredentialsId', env: ['PIPER_configurationUsername', 'PIPER_configurationPassword']],
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
            // get step configuration to access `reportName` & `reportFile` & `qualityGates`
            Map stepConfig = readJSON(text: sh(returnStdout: true, script: "${piperGoPath} getConfig --stepMetadata '.pipeline/tmp/${METADATA_FILE}'${customDefaultConfig}${customConfigArg}"))
            echo "Step Config: ${stepConfig}"

            piperExecuteBin.dockerWrapper(script, STEP_NAME, config){
                piperExecuteBin.handleErrorDetails(STEP_NAME) {
                    script.commonPipelineEnvironment.writeToDisk(script)
                    issuesWrapper(parameters, script){
                        try{
                            piperExecuteBin.credentialWrapper(config, credentialInfo){
                                sh "${piperGoPath} ${STEP_NAME}${customDefaultConfig}${customConfigArg}"
                            }
                        } finally {
                            jenkinsUtils.handleStepResults(STEP_NAME, true, false)
                            script.commonPipelineEnvironment.readFromDisk(script)
                        }
                    }
                }
            }
        }
    }
}

def issuesWrapper(Map parameters = [:], Script script, Closure body){
    String stageName = parameters.stageName ?: env.STAGE_NAME
    // load default & individual configuration
    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults([:], stageName)
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .use()

    try {
        body()
    } finally {
        recordIssues(
            blameDisabled: true,
            enabledForFailure: true,
            aggregatingResults: false,
            qualityGates: config.qualityGates,
            tool: checkStyle(
                id: config.reportName,
                name: config.reportName,
                pattern: config.reportFile
            )
        )
    }
}
