import com.sap.piper.GitUtils
import groovy.transform.Field

import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException


@Field def STEP_NAME = 'transportRequestUploadFile'

@Field Set generalConfigurationKeys = [
    'changeManagement',
  ]

@Field Set parameterKeys = generalConfigurationKeys.plus([
    'applicationId',
    'changeDocumentId',
    'filePath',
    'transportRequestId'])

@Field Set stepConfigurationKeys = generalConfigurationKeys

def call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        ChangeManagement cm = parameters.cmUtils ?: new ChangeManagement(script)

        ConfigurationHelper configHelper =
            ConfigurationHelper.loadStepDefaults(this)
                               .mixinGeneralConfig(script.commonPipelineEnvironment, generalConfigurationKeys)
                               .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, stepConfigurationKeys)
                               .mixinStepConfig(script.commonPipelineEnvironment, stepConfigurationKeys)
                               .mixin(parameters, parameterKeys)
                               .withMandatoryProperty('applicationId')
                               .withMandatoryProperty('changeManagement/changeDocumentLabel')
                               .withMandatoryProperty('changeManagement/clientOpts')
                               .withMandatoryProperty('changeManagement/credentialsId')
                               .withMandatoryProperty('changeManagement/endpoint')
                               .withMandatoryProperty('changeManagement/git/from')
                               .withMandatoryProperty('changeManagement/git/to')
                               .withMandatoryProperty('changeManagement/git/format')
                               .withMandatoryProperty('filePath')

        Map configuration = configHelper.use()

        def changeDocumentId = configuration.changeDocumentId

        if(changeDocumentId?.trim()) {

          echo "[INFO] ChangeDocumentId '${changeDocumentId}' retrieved from parameters."

        } else {

          echo "[INFO] Retrieving ChangeDocumentId from commit history [from: ${configuration.changeManagement.git.from}, to: ${configuration.changeManagement.git.to}]." +
               "Searching for pattern '${configuration.changeManagement.changeDocumentLabel}'. Searching with format '${configuration.changeManagement.git.format}'."

            try {
                changeDocumentId = cm.getChangeDocumentId(
                                                  configuration.changeManagement.git.from,
                                                  configuration.changeManagement.git.to,
                                                  configuration.changeManagement.changeDocumentLabel,
                                                  configuration.changeManagement.git.format
                                                 )

                echo "[INFO] ChangeDocumentId '${changeDocumentId}' retrieved from commit history"

            } catch(ChangeManagementException ex) {
                echo "[WARN] Cannot retrieve changeDocumentId from commit history: ${ex.getMessage()}."
            }
        }

        def transportRequestId = configuration.transportRequestId

        if(transportRequestId?.trim()) {

          echo "[INFO] Transport request id '${transportRequestId}' retrieved from parameters."

        } else {

          echo "[INFO] Retrieving transport request id from commit history [from: ${configuration.changeManagement.git.from}, to: ${configuration.changeManagement.git.to}]." +
               " Searching for pattern '${configuration.changeManagement.transportRequestLabel}'. Searching with format '${configuration.changeManagement.git.format}'."

            try {
                transportRequestId = cm.getTransportRequestId(
                                                  configuration.changeManagement.git.from,
                                                  configuration.changeManagement.git.to,
                                                  configuration.changeManagement.transportRequestLabel,
                                                  configuration.changeManagement.git.format
                                                 )

                echo "[INFO] Transport request id '${transportRequestId}' retrieved from commit history"

            } catch(ChangeManagementException ex) {
                echo "[WARN] Cannot retrieve transportRequestId from commit history: ${ex.getMessage()}."
            }
        }

        configuration = configHelper
                           .mixin([changeDocumentId: changeDocumentId?.trim() ?: null,
                                   transportRequestId: transportRequestId?.trim() ?: null], ['changeDocumentId', 'transportRequestId'] as Set)
                           .withMandatoryProperty('changeDocumentId',
                               "Change document id not provided (parameter: \'changeDocumentId\' or via commit history).")
                           .withMandatoryProperty('transportRequestId',
                               "Transport request id not provided (parameter: \'transportRequestId\' or via commit history).")
                           .use()

        echo "[INFO] Uploading file '${configuration.filePath}' to transport request '${configuration.transportRequestId}' of change document '${configuration.changeDocumentId}'."

        withCredentials([usernamePassword(
            credentialsId: configuration.changeManagement.credentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {

            try {
                cm.uploadFileToTransportRequest(configuration.changeDocumentId,
                                                configuration.transportRequestId,
                                                configuration.applicationId,
                                                configuration.filePath,
                                                configuration.changeManagement.endpoint,
                                                username,
                                                password,
                                                configuration.changeManagement.clientOpts)
            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }
        }

        echo "[INFO] File '${configuration.filePath}' has been successfully uploaded to transport request '${configuration.transportRequestId}' of change document '${configuration.changeDocumentId}'."
    }
}
