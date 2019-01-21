import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import com.sap.piper.GitUtils
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

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

@Field Set STEP_CONFIG_KEYS = [
    'changeManagement',
    /**
      * When set to `false` the step will not fail in case the step is not in status 'in development'.
      * @possibleValues `true`, `false`
      */
    'failIfStatusIsNotInDevelopment'
  ]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus('changeDocumentId')

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

        GitUtils gitUtils = parameters?.gitUtils ?: new GitUtils()

        ChangeManagement cm = parameters?.cmUtils ?: new ChangeManagement(script, gitUtils)

        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)

        Map configuration =  configHelper.use()

        BackendType backendType = getBackendTypeAndLogInfoIfCMIntegrationDisabled(this, configuration)
        if(backendType == BackendType.NONE) return

        configHelper
            // for the following parameters we expect defaults
            /**
              * A pattern used for identifying lines holding the change document id.
              * @possibleValues regex pattern
              */
            .withMandatoryProperty('changeManagement/changeDocumentLabel')
            /**
              * Additional options for cm command line client, e.g. like JAVA_OPTS.
              */
            .withMandatoryProperty('changeManagement/clientOpts')
            /**
              * The id of the credentials to connect to the Solution Manager. The credentials needs to be maintained on Jenkins.
              */
            .withMandatoryProperty('changeManagement/credentialsId')
            /**
              * The starting point for retrieving the change document id
              */
            .withMandatoryProperty('changeManagement/git/from')
            /**
              *  The end point for retrieving the change document id
              */
            .withMandatoryProperty('changeManagement/git/to')
            /**
              * Specifies what part of the commit is scanned. By default the body of the commit message is scanned.
              * @possibleValues see `git log --help`
              */
            .withMandatoryProperty('changeManagement/git/format')
            .withMandatoryProperty('failIfStatusIsNotInDevelopment')
            // for the following parameters we expect a value provided from outside
            /**
              * The address of the Solution Manager.
              */
            .withMandatoryProperty('changeManagement/endpoint')

        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], configuration)

        def changeId = getChangeDocumentId(cm, script, configuration)

        configuration = configHelper.mixin([changeDocumentId: changeId?.trim() ?: null], ['changeDocumentId'] as Set)

                                    /**
                                      * The id of the change document to transport. If not provided, it is retrieved from the git commit history.
                                      */
                                    .withMandatoryProperty('changeDocumentId',
                                        "No changeDocumentId provided. Neither via parameter 'changeDocumentId' " +
                                        "nor via label '${configuration.changeManagement.changeDocumentLabel}' in commit range " +
                                        "[from: ${configuration.changeManagement.git.from}, to: ${configuration.changeManagement.git.to}].")
                                    .use()

        boolean isInDevelopment

        echo "[INFO] Checking if change document '${configuration.changeDocumentId}' is in development."

        try {

            isInDevelopment = cm.isChangeInDevelopment(configuration.changeDocumentId,
                configuration.changeManagement.endpoint,
                configuration.changeManagement.credentialsId,
                configuration.changeManagement.clientOpts)

        } catch(ChangeManagementException ex) {
            throw new AbortException(ex.getMessage())
        }


        if(isInDevelopment) {
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
