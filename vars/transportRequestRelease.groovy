import com.sap.piper.GitUtils
import groovy.transform.Field

import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException


@Field def STEP_NAME = 'transportRequestRelease'

@Field Set parameterKeys = [
    'changeDocumentId',
    'transportRequestId',
    'credentialsId',
    'endpoint',
    'gitChangeDocumentLabel',
    'gitFormat',
    'gitFrom',
    'gitTo'
  ]

@Field Set stepConfigurationKeys = [
    'credentialsId',
    'endpoint',
    'gitChangeDocumentLabel',
    'gitFormat',
    'gitFrom',
    'gitTo'
  ]

def call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        ChangeManagement cm = parameters.cmUtils ?: new ChangeManagement(script)

        Map configuration = ConfigurationMerger.merge(script, STEP_NAME,
                                                      parameters, parameterKeys,
                                                      stepConfigurationKeys)

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

        if(!changeDocumentId) {
            throw new AbortException("Change document id not provided (parameter: 'changeDocumentId' or via commit history).")
        }

        boolean isInDevelopment

        def transportRequestId = configuration.transportRequestId
        if(!transportRequestId) throw new AbortException("Transport Request id not provided (parameter: 'transportRequestId').")

        def credentialsId = configuration.credentialsId
        if(!credentialsId) throw new AbortException("Credentials id not provided (parameter: 'credentialsId').")

        def endpoint = configuration.endpoint
        if(!endpoint) throw new AbortException("Solution Manager endpoint not provided (parameter: 'endpoint').")

        echo "[INFO] Closing transport request '$transportRequestId' for change document '$changeDocumentId'."

        withCredentials([usernamePassword(
            credentialsId: credentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {

            try {
                cm.releaseTransportRequest(changeDocumentId, transportRequestId, endpoint, username, password)
            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }
        }

        echo "[INFO] Transport Request '${transportRequestId}' has been successfully closed."
    }
}
