import static com.sap.piper.Prerequisites.checkScript
import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Name of the docker image that should be used, in which node should be installed and configured. Default value is 'hadolint/hadolint:latest-debian'.
     */
    'dockerImage'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Name of the configuration file used locally within the step.
     */
    'configurationFile',
    /**
     * URL pointing to the .hadolint.yaml exclude configuration to be used for linting
     */
    'configurationUrl',
    /**
     * Name of the result file used locally within the step.
     */
    'reportFile'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS + [
    /**
     * Docker options to be set when starting the container.
     */
    'dockerOptions'
]
/**
 * Executes the Haskell Dockerfile Linter which is a smarter Dockerfile linter that helps you build [best practice](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/) Docker images.
 * The linter is parsing the Dockerfile into an abstract syntax tree (AST) and performs rules on top of the AST.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        final script = checkScript(this, parameters) ?: this
        final utils = parameters.juStabUtils ?: new Utils()

        // load default & individual configuration
        Map configuration = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], configuration)

        configuration.stashContent = utils.unstashAll(configuration.stashContent)

        if (!fileExists(configuration.dockerfile)) {
            error "[${STEP_NAME}] Dockerfile is not found."
        }

        if(!fileExists(configuration.configurationFile) && configuration.configurationUrl) {
            sh "curl --location --output ${configuration.configurationFile} ${configuration.configurationUrl}"
            if(configuration.stashContent) {
                def stashName = 'hadolintConfiguration'
                stash name: stashName, includes: configuration.configurationFile
                configuration.stashContent += stashName
            }
        }

        def options = [
            "--config ${configuration.configurationFile}",
            "--format checkstyle > ${configuration.reportFile}"
        ]

        dockerExecute(
            script: script,
            dockerImage: configuration.dockerImage,
            dockerOptions: configuration.dockerOptions,
            stashContent: configuration.stashContent
        ) {
            def ignored = sh returnStatus: true, script: "hadolint ${configuration.dockerfile} ${options.join(' ')}"

            recordIssues(
                tools: [checkStyle(name: configuration.reportName, pattern: configuration.reportFile)],
                enabledForFailure: true
            )
            archiveArtifacts configuration.reportFile
        }
    }
}
