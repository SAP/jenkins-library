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

    String getChangeDocumentId(Map config) {

            if(config.changeDocumentId) {
                script.echo "[INFO] Use changeDocumentId '${config.changeDocumentId}' from configuration."
                return config.changeDocumentId
            }

            script.echo "[INFO] Retrieving changeDocumentId from git commit(s) [FROM: ${config.git_from}, TO: ${config.git_to}]"
            def changeDocumentId = getChangeDocumentId(
                                        config.git_from,
                                        config.git_to,
                                        config.git_label,
                                        config.git_format
                                   )
            script.echo "[INFO] ChangeDocumentId '${changeDocumentId}' retrieved from git commit(s)."

            return changeDocumentId
        }

    String getTransportRequestId(def commonPipelineEnvironment,
                                 String transportRequestId,
                                 String label,
                                 String gitFrom,
                                 String gitTo,
                                 String gitFormat
                                 ) {

        String _transportRequestId = transportRequestId

        if(_transportRequestId) {
            commonPipelineEnvironment.setTransportRequestId(_transportRequestId)
        }

        _transportRequestId = commonPipelineEnvironment.getTransportRequestId()

        if(_transportRequestId) {
            return _transportRequestId
        }

        try {
            script.echo "[INFO] Retrieving transportRequestId from git commit(s) [FROM: ${gitFrom}, TO: ${gitTo}]"

            _transportRequestId = getTransportRequestId(
                           label,
                           gitFrom,
                           gitTo,
                           gitFormat
                         )

            script.echo "[INFO] Transport request id '${_transportRequestId}' retrieved from git commit(s)."
            commonPipelineEnvironment.setTransportRequestId(_transportRequestId)

        } catch(ChangeManagementException e) {
            script.echo "[WARN] Cannot retrieve transport request id from commit history."
            // OK, we return null
        }
        return _transportRequestId
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
                              String label = 'TransportRequest\\s?:',
                              String from = 'origin/master',
                              String to = 'HEAD',
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

    boolean isChangeInDevelopment(String changeId, String endpoint, String username, String password, String cmclientOpts = '') {

                int rc = script.sh(returnStatus: true,
                            script: getCMCommandLine(endpoint, username, password,
                                                     'is-change-in-development', ['-cID', "'${changeId}'",
                                                                                   '--return-code'],
                                                                               cmclientOpts))

                if(rc == 0) {
                    return true
                } else if(rc == 3) {
                    return false
                } else {
                    throw new ChangeManagementException("Cannot retrieve status for change document '${changeId}'. Does this change exist? Return code from cmclient: ${rc}.")
                }
            }

    String createTransportRequest(String changeId, String developmentSystemId, String endpoint, String username, String password) {

        try {
          String transportRequest = script.sh(returnStdout: true,
                    script: getCMCommandLine(endpoint, username, password, 'create-transport', ['-cID', changeId,
                                                                                                '-dID', developmentSystemId]))
          return transportRequest.trim()
        } catch(AbortException e) {
          throw new ChangeManagementException("Cannot create a transport request for change id '$changeId'. $e.message.")
        }
    }

    void uploadFileToTransportRequest(String changeId, String transportRequestId, String applicationId, String filePath, String endpoint, String username, String password) {

        int rc = script.sh(returnStatus: true,
                    script: getCMCommandLine(endpoint, username, password,
                                            'upload-file-to-transport', ['-cID', changeId,
                                                                         '-tID', transportRequestId,
                                                                         applicationId, filePath]))

        if(rc == 0) {
            return
        } else {
            throw new ChangeManagementException("Cannot upload file '$filePath' for change document '$changeId' with transport request '$transportRequestId'. Return code from cmclient: $rc.")
        }
    }

    void releaseTransportRequest(String changeId, String transportRequestId, String endpoint, String username, String password) {

        int rc = script.sh(returnStatus: true,
                    script: getCMCommandLine(endpoint, username, password,
                                            'release-transport', ['-cID', changeId,
                                                                  '-tID', transportRequestId]))

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
                            String cmclientOpts = '') {
        String cmCommandLine = '#!/bin/bash'
        if(cmclientOpts) {
            cmCommandLine +=  """
                             export CMCLIENT_OPTS="${cmclientOpts}" """
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
