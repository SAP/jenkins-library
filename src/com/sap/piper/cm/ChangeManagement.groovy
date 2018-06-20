package com.sap.piper.cm

import hudson.AbortException


public class ChangeManagement implements Serializable {

    private script

    public ChangeManagement(def script) {
        this.script = script
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
}
