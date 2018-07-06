import com.sap.piper.GitUtils
import groovy.transform.Field
import hudson.AbortException

import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

@Field def STEP_NAME = 'checkChangeInDevelopment'

@Field Set parameterKeys = [
    'changeDocumentId',
    'cmClientOpts',
    'credentialsId',
    'endpoint',
    'failIfStatusIsNotInDevelopment',
    'gitFrom',
    'gitTo',
    'gitChangeDocumentLabel',
    'gitFormat'
  ]

@Field Set stepConfigurationKeys = [
    'changeDocumentId',
    'cmClientOpts',
    'credentialsId',
    'endpoint',
    'failIfStatusIsNotInDevelopment',
    'gitFrom',
    'gitTo',
    'gitChangeDocumentLabel',
    'gitFormat'
  ]

@Field Set generalConfigurationKeys = stepConfigurationKeys

def call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = parameters.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        GitUtils gitUtils = parameters?.gitUtils ?: new GitUtils()

        ChangeManagement cm = parameters?.cmUtils ?: new ChangeManagement(script, gitUtils)

        Map configuration = ConfigurationHelper
                            .loadStepDefaults(this)
                            .mixinGeneralConfig(script.commonPipelineEnvironment, generalConfigurationKeys)
                            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, stepConfigurationKeys)
                            .mixinStepConfig(script.commonPipelineEnvironment, stepConfigurationKeys)
                            .mixin(parameters, parameterKeys)
                            .use()

        def changeId = configuration.changeDocumentId

        if(changeId?.trim()) {

          echo "[INFO] ChangeDocumentId retrieved from parameters."

        } else {

          echo "[INFO] Retrieving ChangeDocumentId from commit history [from: ${configuration.gitFrom}, to: ${configuration.gitTo}]." +
               "Searching for pattern '${configuration.gitChangeDocumentLabel}'. Searching with format '${configuration.gitFormat}'."

            try {
                changeId = cm.getChangeDocumentId(
                                                  configuration.gitFrom,
                                                  configuration.gitTo,
                                                  configuration.gitChangeDocumentLabel,
                                                  configuration.gitFormat
                                                 )
                if(changeId?.trim()) {
                    echo "[INFO] ChangeDocumentId '${changeId}' retrieved from commit history"
                } else {
                    throw new ChangeManagementException("ChangeId is null or empty.")
                }
            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }
        }

        boolean isInDevelopment

        echo "[INFO] Checking if change document '$changeId' is in development."

        withCredentials([usernamePassword(
            credentialsId: configuration.credentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {

            try {
                isInDevelopment = cm.isChangeInDevelopment(changeId, configuration.endpoint, username, password, configuration.cmClientOpts)
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
