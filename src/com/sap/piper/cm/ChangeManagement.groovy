package com.sap.piper.cm

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
        int rc = executeWithCredentials(BackendType.SOLMAN, endpoint, credentialsId, 'is-change-in-development', ['-cID', "'${changeId}'", '--return-code'],
            false,
            clientOpts) as int

        if (rc == 0) {
            return true
        } else if (rc == 3) {
            return false
        } else {
            throw new ChangeManagementException("Cannot retrieve status for change document '${changeId}'. Does this change exist? Return code from cmclient: ${rc}.")
        }
    }

    String createTransportRequestCTS(String transportType, String targetSystemId, String description, String endpoint, String credentialsId, String clientOpts = '') {
        try {
            def transportRequest = executeWithCredentials(BackendType.CTS, endpoint, credentialsId, 'create-transport',
                    ['-tt', transportType, '-ts', targetSystemId, '-d', "\"${description}\""],
                    true,
                    clientOpts)
            return (transportRequest as String)?.trim()
        }catch(AbortException e) {
            throw new ChangeManagementException("Cannot create a transport request. $e.message.")
        }
    }

    String createTransportRequestSOLMAN(String changeId, String developmentSystemId, String endpoint, String credentialsId, String clientOpts = '') {

        try {
            def transportRequest = executeWithCredentials(BackendType.SOLMAN, endpoint, credentialsId, 'create-transport', ['-cID', changeId, '-dID', developmentSystemId],
                true,
                clientOpts)
            return (transportRequest as String)?.trim()
        }catch(AbortException e) {
            throw new ChangeManagementException("Cannot create a transport request for change id '$changeId'. $e.message.")
        }
    }

    void uploadFileToTransportRequest(BackendType type, String changeId, String transportRequestId, String applicationId, String filePath, String endpoint, String credentialsId, String cmclientOpts = '') {

        def args = null

        if(type == BackendType.SOLMAN) {
            args = ['-cID', changeId,
                    '-tID', transportRequestId,
                    applicationId, "\"$filePath\""]
        } else if (type == BackendType.CTS) {
            args = ['-tID', transportRequestId,
                    "\"$filePath\""]
        } else {
            throw new IllegalArgumentException("Invalid backend type: ${type}")
        }

        int rc = executeWithCredentials(type,
                                        endpoint,
                                        credentialsId,
                                        'upload-file-to-transport',
                                        args,
                                        false,
                                        cmclientOpts) as int

        if(rc == 0) {
            return
        } else {
            throw new ChangeManagementException("Cannot upload file '$filePath' for change document '$changeId' with transport request '$transportRequestId'. Return code from cmclient: $rc.")
        }

    }

    def executeWithCredentials(BackendType type, String endpoint, String credentialsId, String command, List<String> args, boolean returnStdout = false, String clientOpts = '') {
        script.withCredentials([script.usernamePassword(
            credentialsId: credentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {
            def cmScript = getCMCommandLine(type, endpoint, script.username, script.password,
                    command, args,
                    clientOpts)

            Map shArgs = [:]
            if(returnStdout)
                shArgs.put('returnStdout', true)
            else
                shArgs.put('returnStatus', true)

            shArgs.put('script', cmScript)

            // user and password are masked by withCredentials
            script.echo """[INFO] Executing command line: "${cmScript}"."""
            return script.sh(shArgs)
        }
    }

    void releaseTransportRequest(BackendType type,String changeId, String transportRequestId, String endpoint, String credentialsId, String clientOpts = '') {

        def cmd
        List args = []

        if(type == BackendType.SOLMAN) {
            cmd = 'release-transport'
            args << '-cID'
            args << changeId
        } else if(type == BackendType.CTS) {
             cmd = 'export-transport'
        } else {
            throw new IllegalStateException("Invalid backend type: '${type}'")
        }

        args << '-tID'
        args << transportRequestId

        int rc = executeWithCredentials(type, endpoint, credentialsId, cmd, args, false, clientOpts) as int
        if(rc == 0) {
            return
        } else {
            throw new ChangeManagementException("Cannot release Transport Request '$transportRequestId'. Return code from cmclient: $rc.")
        }
    }

    String getCMCommandLine(BackendType type,
                            String endpoint,
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
                           -t ${type} \
                          ${command} ${(args as Iterable).join(' ')}
                    """
        return cmCommandLine
    }
}
