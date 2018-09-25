import com.sap.piper.GitUtils
import com.sap.piper.Utils
import groovy.transform.Field

import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.BackendType
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException


@Field def STEP_NAME = 'transportRequestRelease'

@Field Set stepConfigurationKeys = [
    'changeManagement'
  ]

@Field Set parameterKeys = stepConfigurationKeys.plus([
    'changeDocumentId',
    'transportRequestId',
  ])

@Field Set generalConfigurationKeys = stepConfigurationKeys

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
            .withMandatoryProperty('changeManagement/clientOpts')
            .withMandatoryProperty('changeManagement/credentialsId')
            .withMandatoryProperty('changeManagement/endpoint')
            .withMandatoryProperty('changeManagement/git/to')
            .withMandatoryProperty('changeManagement/git/from')
            .withMandatoryProperty('changeManagement/git/format')

        Map configuration = configHelper.use()

        new Utils().pushToSWA([step: STEP_NAME], configuration)

        BackendType backendType

        try {
            backendType = configuration.changeManagement.type as BackendType
        } catch(IllegalArgumentException e) {
            error "Invalid backend type: '${configuration.changeManagement.type}'. " +
                  "Valid values: [${BackendType.values().join(', ')}]. " +
                  "Configuration: 'changeManagement/type'."
        }

        def transportRequestId = configuration.transportRequestId

        if(transportRequestId?.trim()) {

          echo "[INFO] Transport request id '${transportRequestId}' retrieved from parameters."

        } else {

          echo "[INFO] Retrieving transport request id from commit history [from: ${configuration.changeManagement.git.from}, to: ${configuration.changeManagement.git.to}]." +
               " Searching for pattern '${configuration.gitTransportRequestLabel}'. Searching with format '${configuration.changeManagement.git.format}'."

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

        def changeDocumentId = null

        if(backendType == BackendType.SOLMAN) {

            changeDocumentId = configuration.changeDocumentId

            if(changeDocumentId?.trim()) {

                echo "[INFO] ChangeDocumentId '${changeDocumentId}' retrieved from parameters."

            } else {

                echo "[INFO] Retrieving ChangeDocumentId from commit history [from: ${configuration.changeManagement.git.from}, to: ${configuration.changeManagement.git.to}]." +
                     "Searching for pattern '${configuration.changeDocumentLabel}'. Searching with format '${configuration.changeManagement.git.format}'."

                try {
                    changeDocumentId = cm.getChangeDocumentId(
                                                      configuration.changeManagement.git.from,
                                                      configuration.changeManagement.git.to,
                                                      configuration.changeManagement.changeDocumentLabel,
                                                      configuration.changeManagement.gitformat
                                                     )

                    echo "[INFO] ChangeDocumentId '${changeDocumentId}' retrieved from commit history"

                } catch(ChangeManagementException ex) {
                    echo "[WARN] Cannot retrieve changeDocumentId from commit history: ${ex.getMessage()}."
                }
            }

            configHelper.mixin([changeDocumentId: changeDocumentId?.trim() ?: null], ['changeDocumentId'] as Set)
                        .withMandatoryProperty('changeDocumentId',
                            "Change document id not provided (parameter: \'changeDocumentId\' or via commit history).")

        }

        configuration = configHelper
                            .mixin([transportRequestId: transportRequestId?.trim() ?: null], ['transportRequestId'] as Set)
                            .withMandatoryProperty('transportRequestId',
                                "Transport request id not provided (parameter: \'transportRequestId\' or via commit history).")
                            .use()

        def closingMessage = ["[INFO] Closing transport request '${configuration.transportRequestId}'"]
        if(backendType == BackendType.SOLMAN) closingMessage << " for change document '${configuration.changeDocumentId}'"
        closingMessage << '.'
        echo closingMessage.join()

            try {
                cm.releaseTransportRequest(backendType,
                                           configuration.changeDocumentId,
                                           configuration.transportRequestId,
                                           configuration.changeManagement.endpoint,
                                           configuration.changeManagement.credentialsId,
                                           configuration.changeManagement.clientOpts)

            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }


        echo "[INFO] Transport Request '${configuration.transportRequestId}' has been successfully closed."
    }
}
