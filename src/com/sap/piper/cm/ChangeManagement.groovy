package com.sap.piper.cm

import java.util.Map

import com.sap.piper.GitUtils

import hudson.AbortException


public class ChangeManagement implements Serializable {

    private script
    private GitUtils gitUtils

    public ChangeManagement(def script, GitUtils gitUtils = null) {
        this.script = script
        this.gitUtils = gitUtils ?: new GitUtils()
    }

    String getChangeDocumentId(
                              String from = 'origin/master',
                              String to = 'HEAD',
                              String label = 'ChangeDocument\\s?:',
                              String format = '%b'
                            ) {

        return getLabeledItem('ChangeDocumentId', from, to, label, format)
    }

    String getTransportRequestId(
                              String from = 'origin/master',
                              String to = 'HEAD',
                              String label = 'TransportRequest\\s?:',
                              String format = '%b'
                            ) {

        return getLabeledItem('TransportRequestId', from, to, label, format)
    }

    private String getLabeledItem(
                              String name,
                              String from,
                              String to,
                              String label,
                              String format
                            ) {

        if( ! gitUtils.insideWorkTree() ) {
            throw new ChangeManagementException("Cannot retrieve ${name}. Not in a git work tree. ${name} is extracted from git commit messages.")
        }

        def items = gitUtils.extractLogLines(".*${label}.*", from, to, format)
                                .collect { line -> line?.replaceAll(label,'')?.trim() }
                                .unique()

        items.retainAll { line -> line != null && ! line.isEmpty() }

        if( items.size() == 0 ) {
            throw new ChangeManagementException("Cannot retrieve ${name} from git commits. ${name} retrieved from git commit messages via pattern '${label}'.")
        } else if (items.size() > 1) {
            throw new ChangeManagementException("Multiple ${name}s found: ${items}. ${name} retrieved from git commit messages via pattern '${label}'.")
        }

        return items[0]
    }

    boolean isChangeInDevelopment(String changeId, String endpoint, String credentialsId, String clientOpts = '') {
        int rc = executeWithCredentials(endpoint, credentialsId, 'is-change-in-development', ['-cID', "'${changeId}'", '--return-code'],
            clientOpts) as int

        if (rc == 0) {
            return true
        } else if (rc == 3) {
            return false
        } else {
            throw new ChangeManagementException("Cannot retrieve status for change document '${changeId}'. Does this change exist? Return code from cmclient: ${rc}.")
        }
    }

    String createTransportRequest(String changeId, String developmentSystemId, String endpoint, String credentialsId, String clientOpts = '') {
        try {
            def transportRequest = executeWithCredentials(endpoint, credentialsId, 'create-transport', ['-cID', changeId, '-dID', developmentSystemId],
                clientOpts)
            return transportRequest.trim() as String
        }catch(AbortException e) {
            throw new ChangeManagementException("Cannot create a transport request for change id '$changeId'. $e.message.")
        }
    }


    void uploadFileToTransportRequest(String changeId, String transportRequestId, String applicationId, String filePath, String endpoint, String credentialsId, String cmclientOpts = '') {
        int rc = executeWithCredentials(endpoint, credentialsId, 'upload-file-to-transport', ['-cID', changeId,
                                                                                                 '-tID', transportRequestId,
                                                                                                 applicationId, "\"$filePath\""],
            cmclientOpts) as int

        if(rc == 0) {
            return
        } else {
            throw new ChangeManagementException("Cannot upload file '$filePath' for change document '$changeId' with transport request '$transportRequestId'. Return code from cmclient: $rc.")
        }

    }

    def executeWithCredentials(String endpoint, String credentialsId, String command, List<String> args, String clientOpts = '') {
        script.withCredentials([script.usernamePassword(
            credentialsId: credentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {
            def cmScript = getCMCommandLine(endpoint, script.username, script.password,
                    command, args,
                    clientOpts)
            // user and password are masked by withCredentials
            script.echo """[INFO] Executing command line: "${cmScript}"."""
            def returnValue = script.sh(returnStatus: true,
                script: cmScript)
            return returnValue;

        }

    }

    void releaseTransportRequest(String changeId, String transportRequestId, String endpoint, String credentialsId, String clientOpts = '') {
        int rc = executeWithCredentials( endpoint, credentialsId, 'release-transport', ['-cID', changeId,
                                                                                        '-tID', transportRequestId], clientOpts) as int
        if(rc == 0) {
            return
        } else {
            throw new ChangeManagementException("Cannot release Transport Request '$transportRequestId'. Return code from cmclient: $rc.")
        }
    }

    String getCMCommandLine(String endpoint,
                            String username,
                            String password,
                            String command,
                            List<String> args,
                            String clientOpts = '') {
        String cmCommandLine = '#!/bin/bash'
        if(clientOpts) {
            cmCommandLine +=  """
                             export CMCLIENT_OPTS="${clientOpts}" """
        }
        cmCommandLine += """
                        cmclient -e '$endpoint' \
                           -u '$username' \
                           -p '$password' \
                           -t SOLMAN \
                          ${command} ${(args as Iterable).join(' ')}
                    """
        return cmCommandLine
    }
}
