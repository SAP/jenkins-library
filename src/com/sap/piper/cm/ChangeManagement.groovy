package com.sap.piper.cm

import com.sap.piper.GitUtils

import hudson.AbortException


public class ChangeManagement implements Serializable {

    private script
    private GitUtils gitUtils

    public ChangeManagement(def script, GitUtils gitUtils = null) {
        this.script = script
        this.gitUtils = gitUtils
    }

    String getChangeDocumentId(
                              String from = 'origin/master',
                              String to = 'HEAD',
                              String label = 'ChangeDocument\\s?:',
                              String format = '%b'
                            ) {

        if( ! gitUtils.insideWorkTree() ) {
            throw new ChangeManagementException('Cannot retrieve change document id. Not in a git work tree. Change document id is extracted from git commit messages.')
        }

        def changeIds = gitUtils.extractLogLines(".*${label}.*", from, to, format)
                                .collect { line -> line.replaceAll(label,'').trim() }
                                .unique()
            changeIds.retainAll { line -> ! line.isEmpty() }

        if( changeIds.size() == 0 ) {
            throw new ChangeManagementException("Cannot retrieve changeId from git commits. Change id retrieved from git commit messages via pattern '${label}'.")
        } else if (changeIds.size() > 1) {
            throw new ChangeManagementException("Multiple ChangeIds found: ${changeIds}. Change id retrieved from git commit messages via pattern '${label}'.")
        }

        return changeIds.get(0)
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
