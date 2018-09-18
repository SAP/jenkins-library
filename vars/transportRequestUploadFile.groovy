import com.sap.piper.GitUtils
import com.sap.piper.Utils
import groovy.transform.Field

import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException

@Field def STEP_NAME = 'transportRequestUploadFile'

@Field Set generalConfigurationKeys = [
    'changeManagement'
  ]

  @Field Set stepConfigurationKeys = generalConfigurationKeys.plus([
      'applicationId'
    ])

@Field Set parameterKeys = stepConfigurationKeys.plus([
    'changeDocumentId',
    'filePath',
    'transportRequestId'])

def call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        ChangeManagement cm = parameters.cmUtils ?: new ChangeManagement(script)

        ConfigurationHelper configHelper = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, generalConfigurationKeys)
            .mixinStepConfig(script.commonPipelineEnvironment, stepConfigurationKeys)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, stepConfigurationKeys)
            .mixin(parameters, parameterKeys)
            .addIfEmpty('filePath', script.commonPipelineEnvironment.getMtarFilePath())
            .withMandatoryProperty('applicationId')
            .withMandatoryProperty('changeManagement/changeDocumentLabel')
            .withMandatoryProperty('changeManagement/clientOpts')
            .withMandatoryProperty('changeManagement/credentialsId')
            .withMandatoryProperty('changeManagement/endpoint')
            .withMandatoryProperty('changeManagement/type')
            .withMandatoryProperty('changeManagement/git/from')
            .withMandatoryProperty('changeManagement/git/to')
            .withMandatoryProperty('changeManagement/git/format')
            .withMandatoryProperty('filePath')

        Map configuration = configHelper.use()

        new Utils().pushToSWA([step: STEP_NAME, stepParam1: configuration.changeManagement.type], configuration)

        ChangeManagement.BackendType backendType

        try {
            backendType = configuration.changeManagement.type as ChangeManagement.BackendType
        } catch(IllegalArgumentException e) {
            error "Invalid backend type: '${configuration.changeManagement.type}'. " +
                  "Valid values: [${ChangeManagement.BackendType.values().join(', ')}]. " +
                  "Configuration: 'changeManagement/type'."
        }

        def changeDocumentId = null

        if(backendType == ChangeManagement.BackendType.SOLMAN) {

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

        if(backendType == ChangeManagement.BackendType.SOLMAN) {
            configHelper
                .withMandatoryProperty('changeDocumentId',
                    "Change document id not provided (parameter: \'changeDocumentId\' or via commit history).")
        }
        configuration = configHelper
                            .withMandatoryProperty('transportRequestId',
                               "Transport request id not provided (parameter: \'transportRequestId\' or via commit history).")
                           .use()

        def uploadingMessage = ["[INFO] Uploading file '${configuration.filePath}' to transport request '${configuration.transportRequestId}'"]
        if(backendType == ChangeManagement.BackendType.SOLMAN)
            uploadingMessage << " of change document '${configuration.changeDocumentId}'"
        uploadingMessage << '.'

        echo uploadingMessage.join()

            try {


                cm.uploadFileToTransportRequest(configuration.changeDocumentId,
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
        if(backendType == ChangeManagement.BackendType.SOLMAN)
            uploadedMessage << " of change document '${configuration.changeDocumentId}'"
        uploadedMessage << '.'
        echo uploadedMessage.join()
    }
}
