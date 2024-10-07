import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils
import groovy.transform.Field
import hudson.AbortException

import com.sap.piper.ConfigurationHelper
import com.sap.piper.cm.BackendType
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import static com.sap.piper.cm.StepHelpers.getChangeDocumentId
import static com.sap.piper.cm.StepHelpers.getBackendTypeAndLogInfoIfCMIntegrationDisabled

@Field def STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    'changeManagement',
        /**
         * A pattern used for identifying lines holding the change document id.
         * @possibleValues regex pattern
         * @parentConfigKey changeManagement
         */
        'changeDocumentLabel',
        /**
         * Additional options for cm command line client, e.g. JAVA_OPTS.
         * @parentConfigKey changeManagement
         */
        'clientOpts',
        /**
         * The id of the credentials to connect to the Solution Manager. The credentials needs to be maintained on Jenkins.
         * @parentConfigKey changeManagement
         */
        'credentialsId',
        /**
         * The service endpoint, e.g. Solution Manager, ABAP System.
         * @parentConfigKey changeManagement
         */
        'endpoint',
        /**
         * The starting point for retrieving the change document id.
         * @parentConfigKey changeManagement
         */
        'git/from',
        /**
         * The end point for retrieving the change document id.
         * @parentConfigKey changeManagement
         */
        'git/to',
        /**
         * Specifies what part of the commit is scanned. By default the body of the commit message is scanned.
         * @possibleValues see `git log --help`
         * @parentConfigKey changeManagement
         */
        'git/format'
]

@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(
    /**
     * When set to `false` the step will not fail in case the step is not in status 'in development'.
     * @possibleValues `true`, `false`
     */
    'failIfStatusIsNotInDevelopment')

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus(
    /**
     * The id of the change document to transport. If not provided, it is retrieved from the git commit history.
     */
    'changeDocumentId'
)

/**
 * Checks if a Change Document in SAP Solution Manager is in status 'in development'. The change document id is retrieved from the git commit history. The change document id
 * can also be provided via parameter `changeDocumentId`. Any value provided as parameter has a higher precedence than a value from the commit history.
 *
 * By default the git commit messages between `origin/master` and `HEAD` are scanned for a line like `ChangeDocument : <changeDocumentId>`. The commit
 * range and the pattern can be configured. For details see 'parameters' table.
 */
@GenerateDocumentation
void call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this
        String stageName = parameters.stageName ?: env.STAGE_NAME

        addPipelineWarning(script, "Deprecation Warning", "The step ${STEP_NAME} is deprecated. Follow the documentation for options.")

        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)

        Map configuration =  configHelper.use()

        BackendType backendType = getBackendTypeAndLogInfoIfCMIntegrationDisabled(this, configuration)
        if(backendType == BackendType.NONE) return

        configHelper
            // for the following parameters we expect defaults
            .withMandatoryProperty('changeManagement/changeDocumentLabel')
            .withMandatoryProperty('changeManagement/clientOpts')
            .withMandatoryProperty('changeManagement/credentialsId')
            .withMandatoryProperty('changeManagement/git/from')
            .withMandatoryProperty('changeManagement/git/to')
            .withMandatoryProperty('changeManagement/git/format')
            .withMandatoryProperty('failIfStatusIsNotInDevelopment')
            // for the following parameters we expect a value provided from outside
            .withMandatoryProperty('changeManagement/endpoint')

        def changeId = getChangeDocumentId(script, configuration)

        configuration = configHelper.mixin([changeDocumentId: changeId?.trim() ?: null], ['changeDocumentId'] as Set)
                                    .withMandatoryProperty('changeDocumentId',
                                        "No changeDocumentId provided. Neither via parameter 'changeDocumentId' " +
                                        "nor via label '${configuration.changeManagement.changeDocumentLabel}' in commit range " +
                                        "[from: ${configuration.changeManagement.git.from}, to: ${configuration.changeManagement.git.to}].")
                                    .use()

        boolean isInDevelopment

        echo "[INFO] Checking if change document '${configuration.changeDocumentId}' is in development."

        Map paramsUpload = [
            script: script,
            changeDocumentId: configuration.changeDocumentId,
            endpoint: configuration.changeManagement.endpoint,
            cmClientOpts: configuration.changeManagement.clientOpts?: [:],
            credentialsId: configuration.changeManagement.credentialsId,
            failIfStatusIsNotInDevelopment: configuration.failIfStatusIsNotInDevelopment
            ]

        paramsUpload = addDockerParams(script, paramsUpload, configuration.changeManagement.solman?.docker)

        isChangeInDevelopment(paramsUpload)

        if(script.commonPipelineEnvironment.getValue('isChangeInDevelopment')) {
            echo "[INFO] Change '${changeId}' is in status 'in development'."
        } else {
            if(configuration.failIfStatusIsNotInDevelopment.toBoolean()) {
                throw new AbortException("Change '${changeId}' is not in status 'in development'.")

            } else {
                echo "[WARNING] Change '${changeId}' is not in status 'in development'. Failing the pipeline has been explicitly disabled."
            }
        }
    }
}

Map addDockerParams(Script script, Map parameters, Map docker) {

    if(docker) {
        if(docker.image) {
            parameters.dockerImage = docker.image
        }
        if(docker.options) {
            parameters.dockerOptions = docker.options
        }
        if(docker.envVars) {
            parameters.dockerEnvVars = docker.envVars
        }
        if(docker.pullImage != null) {
            parameters.put('dockerPullImage', docker.pullImage)
        }
    }
    return parameters
}

static void addPipelineWarning(Script script, String heading, String message) {
    script.echo '[WARNING] ' + message
    script.addBadge(icon: "warning.gif", text: message)

    String html =
        """
            <h2>$heading</h2>
            <p>$message</p>
            """

    script.createSummary(icon: "warning.gif", text: html)
}
