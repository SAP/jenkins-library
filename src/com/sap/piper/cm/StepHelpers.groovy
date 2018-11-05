package com.sap.piper.cm;

import com.cloudbees.groovy.cps.NonCPS

public class StepHelpers {

    @NonCPS
    public static def getTransportRequestId(ChangeManagement cm, def step, Map configuration) {

        def transportRequestId = configuration.transportRequestId

        if(transportRequestId?.trim()) {

            step.echo "[INFO] Transport request id '${transportRequestId}' retrieved from parameters."

        } else {

            step.echo "[INFO] Retrieving transport request id from commit history [from: ${configuration.changeManagement.git.from}, to: ${configuration.changeManagement.git.to}]." +
                        " Searching for pattern '${configuration.changeManagement.transportRequestLabel}'. Searching with format '${configuration.changeManagement.git.format}'."

            try {
                transportRequestId = cm.getTransportRequestId(
                                                                configuration.changeManagement.git.from,
                                                                configuration.changeManagement.git.to,
                                                                configuration.changeManagement.transportRequestLabel,
                                                                configuration.changeManagement.git.format
                                                            )

                step.echo "[INFO] Transport request id '${transportRequestId}' retrieved from commit history"

            } catch(ChangeManagementException ex) {
                step.echo "[WARN] Cannot retrieve transportRequestId from commit history: ${ex.getMessage()}."
            }
        }
        return transportRequestId
    }

    @NonCPS
    public static getChangeDocumentId(ChangeManagement cm, def step, Map configuration) {

        def changeDocumentId = configuration.changeDocumentId

        if(changeDocumentId?.trim()) {

            step.echo "[INFO] ChangeDocumentId '${changeDocumentId}' retrieved from parameters."

        } else {

            step.echo "[INFO] Retrieving ChangeDocumentId from commit history [from: ${configuration.changeManagement.git.from}, to: ${configuration.changeManagement.git.to}]." +
                        "Searching for pattern '${configuration.changeManagement.changeDocumentLabel}'. Searching with format '${configuration.changeManagement.git.format}'."

            try {
                changeDocumentId = cm.getChangeDocumentId(
                                                            configuration.changeManagement.git.from,
                                                            configuration.changeManagement.git.to,
                                                            configuration.changeManagement.changeDocumentLabel,
                                                            configuration.changeManagement.git.format
                                                        )

                step.echo "[INFO] ChangeDocumentId '${changeDocumentId}' retrieved from commit history"

            } catch(ChangeManagementException ex) {
                step.echo "[WARN] Cannot retrieve changeDocumentId from commit history: ${ex.getMessage()}."
            }
        }
        return changeDocumentId
    }

    @NonCPS
    static BackendType getBackendTypeAndLogInfoIfCMIntegrationDisabled(def step, Map configuration) {

        BackendType backendType

        try {
            backendType = configuration.changeManagement.type as BackendType
        } catch(IllegalArgumentException e) {
            step.error "Invalid backend type: '${configuration.changeManagement.type}'. " +
                  "Valid values: [${BackendType.values().join(', ')}]. " +
                  "Configuration: 'changeManagement/type'."
        }

        if (backendType == BackendType.NONE) {
            step.echo "[INFO] Change management integration intentionally switched off. " +
                 "In order to enable it provide 'changeManagement/type with one of " +
                 "[${BackendType.values().minus(BackendType.NONE).join(', ')}] and maintain " +
                 "other required properties like 'endpoint', 'credentialsId'."
        }

        return backendType
    }
}
