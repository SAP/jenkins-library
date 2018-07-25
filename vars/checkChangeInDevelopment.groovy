import com.sap.piper.GitUtils
import groovy.transform.Field
import hudson.AbortException

import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

@Field def STEP_NAME = 'checkChangeInDevelopment'

@Field Set stepConfigurationKeys = [
    'changeManagement',
    'failIfStatusIsNotInDevelopment'
  ]

@Field Set parameterKeys = stepConfigurationKeys.plus('changeDocumentId')

@Field Set generalConfigurationKeys = stepConfigurationKeys

def call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = parameters.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        GitUtils gitUtils = parameters?.gitUtils ?: new GitUtils()

        ChangeManagement cm = parameters?.cmUtils ?: new ChangeManagement(script, gitUtils)

        ConfigurationHelper configHelper = ConfigurationHelper
                                           .loadStepDefaults(this)
                                           .mixinGeneralConfig(script.commonPipelineEnvironment, generalConfigurationKeys)
                                           .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, stepConfigurationKeys)
                                           .mixinStepConfig(script.commonPipelineEnvironment, stepConfigurationKeys)
                                           .mixin(parameters, parameterKeys)
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


        Map configuration = configHelper.use()

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
                                    .withMandatoryProperty('changeDocumentId',
                                        "No changeDocumentId provided. Neither via parameter 'changeDocumentId' " +
                                        "nor via label '${configuration.changeManagement.changeDocumentLabel}' in commit range " +
                                        "[from: ${configuration.changeManagement.git.from}, to: ${configuration.changeManagement.git.to}].")
                                    .use()

        boolean isInDevelopment

        echo "[INFO] Checking if change document '${configuration.changeDocumentId}' is in development."

        withCredentials([usernamePassword(
            credentialsId: configuration.changeManagement.credentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {

            try {
                isInDevelopment = cm.isChangeInDevelopment(configuration.changeDocumentId,
                                                           configuration.changeManagement.endpoint,
                                                           username,
                                                           password,
                                                           configuration.changeManagement.clientOpts)
            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }
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
