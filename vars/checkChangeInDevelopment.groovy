import com.sap.piper.GitUtils
import com.sap.piper.Utils
import groovy.transform.Field
import hudson.AbortException

import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

@Field def STEP_NAME = 'checkChangeInDevelopment'

@Field Set STEP_CONFIG_KEYS = [
    'changeManagement',
    'failIfStatusIsNotInDevelopment'
  ]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus('changeDocumentId')

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

def call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = parameters.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        GitUtils gitUtils = parameters?.gitUtils ?: new GitUtils()

        ChangeManagement cm = parameters?.cmUtils ?: new ChangeManagement(script, gitUtils)

        ConfigurationHelper configHelper = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            // for the following parameters we expect defaults
             /**
               * A pattern used for identifying lines holding the change document id.
               * @value regex pattern
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
               * @value see `git log --help`
               */
            .withMandatoryProperty('changeManagement/git/format')
            /**
              * When set to `false` the step will not fail in case the step is not in status 'in development'.
              * @value `true`, `false`
              */
            .withMandatoryProperty('failIfStatusIsNotInDevelopment')
            // for the following parameters we expect a value provided from outside
            /**
              * The address of the Solution Manager.
              */
            .withMandatoryProperty('changeManagement/endpoint')


        Map configuration = configHelper.use()

        new Utils().pushToSWA([step: STEP_NAME], configuration)

        def changeId = configuration.changeDocumentId

        if(changeId?.trim()) {

            echo "[INFO] ChangeDocumentId retrieved from parameters."

        } else {

          echo "[INFO] Retrieving ChangeDocumentId from commit history [from: ${configuration.changeManagement.git.from}, to: ${configuration.changeManagement.git.to}]." +
               "Searching for pattern '${configuration.changeManagement.changeDocumentLabel}'. Searching with format '${configuration.changeManagement.git.format}'."

            try {
                changeId = cm.getChangeDocumentId(
                                                  configuration.changeManagement.git.from,
                                                  configuration.changeManagement.git.to,
                                                  configuration.changeManagement.changeDocumentLabel,
                                                  configuration.changeManagement.git.format
                                                 )
                if(changeId?.trim()) {
                    echo "[INFO] ChangeDocumentId '${changeId}' retrieved from commit history"
                }
            } catch(ChangeManagementException ex) {
                echo "[WARN] Cannot retrieve changeDocumentId from commit history: ${ex.getMessage()}."
            }
        }

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
            return true
        } else {
            if(configuration.failIfStatusIsNotInDevelopment.toBoolean()) {
                throw new AbortException("Change '${changeId}' is not in status 'in development'.")

            } else {
                echo "[WARNING] Change '${changeId}' is not in status 'in development'. Failing the pipeline has been explicitly disabled."
                return false
            }
        }
    }
}
