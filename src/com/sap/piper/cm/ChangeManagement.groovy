package com.sap.piper.cm

import hudson.AbortException


public class ChangeManagement implements Serializable {

    private script

    public ChangeManagement(def script) {
        this.script = script
    }

    boolean isChangeInDevelopment(String changeId, String endpoint, String username, String password) {

                int rc = script.sh(returnStatus: true,
                            script:
                            """#!/bin/bash
                       cmclient -e "${endpoint}" \
                                -u '${username}' \
                                -p '${password}' \
                                -t SOLMAN\
                              is-change-in-development -cID '${changeId}' \
                                                       --return-code
                    """)

                if(rc == 0) {
                    return true
                } else if(rc == 3) {
                    return false
                } else {
                    throw new ChangeManagementException("Cannot retrieve change status. Return code from cmclient: ${rc}.")
                }
            }

    String createTransportRequest(String changeId, String developmentSystemId, String endpoint, String username, String password) {

        try {
          String transportRequest = script.sh(returnStdout: true,
                    script:
                    """#!/bin/bash
                       cmclient -e '$endpoint' \
                                -u '$username' \
                                -p '$password' \
                                -t SOLMAN \
                              create-transport -cID '$changeId' -dID '$developmentSystemId'
                    """)
          return transportRequest.trim()
        } catch(AbortException e) {
          throw new ChangeManagementException("Cannot create a transport request for change id '$changeId'. $e.message.")
        }
    }

    void uploadFileToTransportRequest(String changeId, String transportRequestId, String applicationId, String filePath, String endpoint, String username, String password) {

        int rc = script.sh(returnStatus: true,
                    script:
                    """#!/bin/bash
                       cmclient -e '$endpoint' \
                                -u '$username' \
                                -p '$password' \
                                -t SOLMAN \
                              upload-file-to-transport -cID '$changeId' -tID '$transportRequestId' '$applicationId' '$filePath'
                    """)

        if(rc == 0) {
            return
        } else {
            throw new ChangeManagementException("Cannot upload file '$filePath' for change document '$changeId' with transport request '$transportRequestId'. Return code from cmclient: $rc.")
        }
    }

    void releaseTransportRequest(String changeId, String transportRequestId, String endpoint, String username, String password) {

        int rc = script.sh(returnStatus: true,
                    script:
                    """#!/bin/bash
                       cmclient -e '$endpoint' \
                                -u '$username' \
                                -p '$password' \
                                -t SOLMAN \
                              release-transport -cID '$changeId' -tID '$transportRequestId'
                    """)

        if(rc == 0) {
            return
        } else {
            throw new ChangeManagementException("Cannot release Transport Request '$transportRequestId'. Return code from cmclient: $rc.")
        }
    }
}
