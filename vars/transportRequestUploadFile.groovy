import com.sap.piper.GitUtils
import groovy.transform.Field

import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException


@Field def STEP_NAME = 'transportRequestUploadFile'

@Field Set parameterKeys = [
    'changeDocumentId',
    'cmClientOpts',
    'transportRequestId',
    'applicationId',
    'filePath',
    'credentialsId',
    'endpoint',
    'gitFrom',
    'gitTo',
    'gitChangeDocumentLabel',
    'gitTransportRequestLabel',
    'gitFormat'
  ]

@Field Set generalConfigurationKeys = [
    'credentialsId',
    'cmClientOpts',
    'endpoint',
    'gitFrom',
    'gitTo',
    'gitChangeDocumentLabel',
    'gitTransportRequestLabel',
    'gitFormat'
  ]

@Field Set stepConfigurationKeys = generalConfigurationKeys

def call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        ChangeManagement cm = parameters.cmUtils ?: new ChangeManagement(script)

        Map configuration = ConfigurationHelper
                            .loadStepDefaults(this)
                            .mixinGeneralConfig(script.commonPipelineEnvironment, generalConfigurationKeys)
                            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, stepConfigurationKeys)
                            .mixinStepConfig(script.commonPipelineEnvironment, stepConfigurationKeys)
                            .mixin(parameters, parameterKeys)
                            .use()

        def changeDocumentId = configuration.changeDocumentId

        if(changeDocumentId?.trim()) {

          echo "[INFO] ChangeDocumentId '${changeDocumentId}' retrieved from parameters."

        } else {

          echo "[INFO] Retrieving ChangeDocumentId from commit history [from: ${configuration.gitFrom}, to: ${configuration.gitTo}]." +
               "Searching for pattern '${configuration.gitChangeDocumentLabel}'. Searching with format '${configuration.gitFormat}'."

            try {
                changeDocumentId = cm.getChangeDocumentId(
                                                  configuration.gitFrom,
                                                  configuration.gitTo,
                                                  configuration.gitChangeDocumentLabel,
                                                  configuration.gitFormat
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

          echo "[INFO] Retrieving transport request id from commit history [from: ${configuration.gitFrom}, to: ${configuration.gitTo}]." +
               " Searching for pattern '${configuration.gitTransportRequestLabel}'. Searching with format '${configuration.gitFormat}'."

            try {
                transportRequestId = cm.getTransportRequestId(
                                                  configuration.gitFrom,
                                                  configuration.gitTo,
                                                  configuration.gitTransportRequestLabel,
                                                  configuration.gitFormat
                                                 )

                echo "[INFO] Transport request id '${transportRequestId}' retrieved from commit history"

            } catch(ChangeManagementException ex) {
                echo "[WARN] Cannot retrieve transportRequestId from commit history: ${ex.getMessage()}."
            }
        }

        if(! changeDocumentId?.trim()) {
            throw new AbortException("Change document id not provided (parameter: 'changeDocumentId' or via commit history).")
        }

        if(!transportRequestId?.trim()) throw new AbortException("Transport Request id not provided (parameter: 'transportRequestId' or via commit history).")

        def applicationId = configuration.applicationId
        if(!applicationId) throw new AbortException("Application id not provided (parameter: 'applicationId').")

        def filePath = configuration.filePath
        if(!filePath) throw new AbortException("File path not provided (parameter: 'filePath').")

        def credentialsId = configuration.credentialsId
        if(!credentialsId) throw new AbortException("Credentials id not provided (parameter: 'credentialsId').")

        def endpoint = configuration.endpoint
        if(!endpoint) throw new AbortException("Solution Manager endpoint not provided (parameter: 'endpoint').")

        echo "[INFO] Uploading file '$filePath' to transport request '$transportRequestId' of change document '$changeDocumentId'."

        withCredentials([usernamePassword(
            credentialsId: credentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {

            try {
                cm.uploadFileToTransportRequest(changeDocumentId, transportRequestId, applicationId, filePath, endpoint, username, password, configuration.cmClientOpts)
            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }
        }

        echo "[INFO] File '$filePath' has been successfully uploaded to transport request '$transportRequestId' of change document '$changeDocumentId'."
    }
}
