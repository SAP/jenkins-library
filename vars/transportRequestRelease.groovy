import com.sap.piper.GitUtils
import groovy.transform.Field

import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException


@Field def STEP_NAME = 'transportRequestRelease'

@Field Set stepConfigurationKeys = [
    'credentialsId',
    'cmClientOpts',
    'endpoint',
    'gitChangeDocumentLabel',
    'gitFrom',
    'gitTo',
    'gitTransportRequestLabel',
    'gitFormat'
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
                            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, stepConfigurationKeys)
                            .mixinStepConfig(script.commonPipelineEnvironment, stepConfigurationKeys)
                            .mixin(parameters, parameterKeys)
                            .withMandatoryProperty('changeDocumentId')
                            .withMandatoryProperty('endpoint')

        Map configuration = configHelper.use()

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

        configuration = configHelper
                            .mixin([transportRequestId: transportRequestId?.trim() ?: null,
                                    changeDocumentId: changeDocumentId?.trim() ?: null], ['transportRequestId', 'changeDocumentId'] as Set)
                            .withMandatoryProperty('transportRequestId',
                                "Transport request id not provided (parameter: \'transportRequestId\' or via commit history).")
                            .withMandatoryProperty('changeDocumentId',
                                "Change document id not provided (parameter: \'changeDocumentId\' or via commit history).")
                            .use()

        echo "[INFO] Closing transport request '${configuration.transportRequestId}' for change document '${configuration.changeDocumentId}'."

        withCredentials([usernamePassword(
            credentialsId: configuration.credentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {

            try {
                cm.releaseTransportRequest(configuration.changeDocumentId,
                                           configuration.transportRequestId,
                                           configuration.endpoint,
                                           username,
                                           password,
                                           configuration.cmClientOpts)
            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }
        }

        echo "[INFO] Transport Request '${configuration.transportRequestId}' has been successfully closed."
    }
}
