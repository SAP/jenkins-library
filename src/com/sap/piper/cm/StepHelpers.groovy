package com.sap.piper.cm;

import com.cloudbees.groovy.cps.NonCPS

public class StepHelpers {

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
