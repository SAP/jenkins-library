import com.sap.piper.Utils
import groovy.transform.Field

import com.sap.piper.ConfigurationHelper
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.BackendType
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException

import static com.sap.piper.cm.StepHelpers.getBackendTypeAndLogInfoIfCMIntegrationDisabled

@Field def STEP_NAME = 'transportRequestUploadFile'

@Field Set GENERAL_CONFIG_KEYS = [
    'changeManagement'
  ]

@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
      'applicationId'
    ])

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    'changeDocumentId',
    'filePath',
    'transportRequestId'])

void call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        ChangeManagement cm = parameters.cmUtils ?: new ChangeManagement(script)

        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .addIfEmpty('filePath', script.commonPipelineEnvironment.getMtarFilePath())

        Map configuration = configHelper.use()

        BackendType backendType = getBackendTypeAndLogInfoIfCMIntegrationDisabled(this, configuration)
        if(backendType == BackendType.NONE) return

        configHelper
            .withMandatoryProperty('changeManagement/changeDocumentLabel')
            .withMandatoryProperty('changeManagement/clientOpts')
            .withMandatoryProperty('changeManagement/credentialsId')
            .withMandatoryProperty('changeManagement/endpoint')
            .withMandatoryProperty('changeManagement/type')
            .withMandatoryProperty('changeManagement/git/from')
            .withMandatoryProperty('changeManagement/git/to')
            .withMandatoryProperty('changeManagement/git/format')
            .withMandatoryProperty('filePath')

        new Utils().pushToSWA([step: STEP_NAME, stepParam1: configuration.changeManagement.type], configuration)

        def changeDocumentId = null

        if(backendType == BackendType.SOLMAN) {

            changeDocumentId = configuration.changeDocumentId

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

        configHelper
            .mixin([changeDocumentId: changeDocumentId?.trim() ?: null,
                    transportRequestId: transportRequestId?.trim() ?: null], ['changeDocumentId', 'transportRequestId'] as Set)

        if(backendType == BackendType.SOLMAN) {
            configHelper
                .withMandatoryProperty('changeDocumentId',
                    "Change document id not provided (parameter: \'changeDocumentId\' or via commit history).")
                .withMandatoryProperty('applicationId')
        }
        configuration = configHelper
                            .withMandatoryProperty('transportRequestId',
                               "Transport request id not provided (parameter: \'transportRequestId\' or via commit history).")
                           .use()

        def uploadingMessage = ["[INFO] Uploading file '${configuration.filePath}' to transport request '${configuration.transportRequestId}'"]
        if(backendType == BackendType.SOLMAN)
            uploadingMessage << " of change document '${configuration.changeDocumentId}'"
        uploadingMessage << '.'

        echo uploadingMessage.join()

            try {


                cm.uploadFileToTransportRequest(backendType,
                                                configuration.changeDocumentId,
                                                configuration.transportRequestId,
                                                configuration.applicationId,
                                                configuration.filePath,
                                                configuration.changeManagement.endpoint,
                                                configuration.changeManagement.credentialsId,
                                                configuration.changeManagement.clientOpts)

            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }


        def uploadedMessage = ["[INFO] File '${configuration.filePath}' has been successfully uploaded to transport request '${configuration.transportRequestId}'"]
        if(backendType == BackendType.SOLMAN)
            uploadedMessage << " of change document '${configuration.changeDocumentId}'"
        uploadedMessage << '.'
        echo uploadedMessage.join()
    }
}
