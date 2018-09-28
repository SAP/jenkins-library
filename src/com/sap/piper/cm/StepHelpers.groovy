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
}
