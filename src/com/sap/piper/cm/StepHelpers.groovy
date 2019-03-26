package com.sap.piper.cm;

import com.cloudbees.groovy.cps.NonCPS

public class StepHelpers {

    public static def getTransportRequestId(ChangeManagement cm, def script, Map configuration) {

        def transportRequestId = configuration.transportRequestId

        if(transportRequestId?.trim()) {

            script.echo "[INFO] Transport request id '${transportRequestId}' retrieved from parameters."
            return transportRequestId

        }

        transportRequestId = script.commonPipelineEnvironment.getCustomProperty('transportRequestId')

        if(transportRequestId?.trim()) {
            script.echo "[INFO] Transport request id '${transportRequestId}' retrieved from common pipeline environment."
            return transportRequestId
        }

        script.echo "[INFO] Retrieving transport request id from commit history [from: ${configuration.changeManagement.git.from}, to: ${configuration.changeManagement.git.to}]." +
                    " Searching for pattern '${configuration.changeManagement.transportRequestLabel}'. Searching with format '${configuration.changeManagement.git.format}'."

        try {
            transportRequestId = cm.getTransportRequestId(
                                                            configuration.changeManagement.git.from,
                                                            configuration.changeManagement.git.to,
                                                            configuration.changeManagement.transportRequestLabel,
                                                            configuration.changeManagement.git.format
                                                        )

            script.commonPipelineEnvironment.setCustomProperty('transportRequestId', "${transportRequestId}")
            script.echo "[INFO] Transport request id '${transportRequestId}' retrieved from commit history"

        } catch(ChangeManagementException ex) {
            script.echo "[WARN] Cannot retrieve transportRequestId from commit history: ${ex.getMessage()}."
        }

        transportRequestId
    }

    public static getChangeDocumentId(ChangeManagement cm, def script, Map configuration) {

        def changeDocumentId = configuration.changeDocumentId

        if(changeDocumentId?.trim()) {

            script.echo "[INFO] ChangeDocumentId '${changeDocumentId}' retrieved from parameters."
            return changeDocumentId
        }

        changeDocumentId = script.commonPipelineEnvironment.getChangeDocumentId()

        if(changeDocumentId?.trim()) {

            script.echo "[INFO] ChangeDocumentId '${changeDocumentId}' retrieved from common pipeline environment."
            return changeDocumentId
        }

        script.echo "[INFO] Retrieving ChangeDocumentId from commit history [from: ${configuration.changeManagement.git.from}, to: ${configuration.changeManagement.git.to}]." +
                    "Searching for pattern '${configuration.changeManagement.changeDocumentLabel}'. Searching with format '${configuration.changeManagement.git.format}'."

        try {
            changeDocumentId = cm.getChangeDocumentId(
                                                        configuration.changeManagement.git.from,
                                                        configuration.changeManagement.git.to,
                                                        configuration.changeManagement.changeDocumentLabel,
                                                        configuration.changeManagement.git.format
                                                    )

            script.echo "[INFO] ChangeDocumentId '${changeDocumentId}' retrieved from commit history"
            script.commonPipelineEnvironment.setChangeDocumentId(changeDocumentId)

        } catch(ChangeManagementException ex) {
            script.echo "[WARN] Cannot retrieve changeDocumentId from commit history: ${ex.getMessage()}."
        }

        return changeDocumentId
    }

    @NonCPS
    static BackendType getBackendTypeAndLogInfoIfCMIntegrationDisabled(def script, Map configuration) {

        BackendType backendType

        try {
            backendType = configuration.changeManagement.type as BackendType
        } catch(IllegalArgumentException e) {
            script.error "Invalid backend type: '${configuration.changeManagement.type}'. " +
                  "Valid values: [${BackendType.values().join(', ')}]. " +
                  "Configuration: 'changeManagement/type'."
        }

        if (backendType == BackendType.NONE) {
            script.echo "[INFO] Change management integration intentionally switched off. " +
                 "In order to enable it provide 'changeManagement/type with one of " +
                 "[${BackendType.values().minus(BackendType.NONE).join(', ')}] and maintain " +
                 "other required properties like 'endpoint', 'credentialsId'."
        }

        return backendType
    }
}
