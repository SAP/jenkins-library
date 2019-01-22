package com.sap.piper.cm

import com.sap.piper.GitUtils

import groovy.json.JsonSlurper
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
        int rc = executeWithCredentials(BackendType.SOLMAN, '', [], endpoint, credentialsId, 'is-change-in-development', ['-cID', "'${changeId}'", '--return-code'],
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
            def transportRequest = executeWithCredentials(BackendType.CTS, '', [], endpoint, credentialsId, 'create-transport',
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
            def transportRequest = executeWithCredentials(BackendType.SOLMAN, '', [], endpoint, credentialsId, 'create-transport', ['-cID', changeId, '-dID', developmentSystemId],
                true,
                clientOpts)
            return (transportRequest as String)?.trim()
        }catch(AbortException e) {
            throw new ChangeManagementException("Cannot create a transport request for change id '$changeId'. $e.message.")
        }
    }

    String createTransportRequestRFC(
        String dockerImage,
        List dockerOptions,
        String endpoint,
        String client,
        String credentialsId,
        String description) {

        def command = 'cts createTransportRequest'
        List args = [
            "--env TRANSPORT_DESCRIPTION=${description}",
            "--env ABAP_DEVELOPMENT_CLIENT=${client}"]

        def transportRequestId = executeWithCredentials(
            BackendType.RFC,
            'rfc',
            dockerOptions,
            endpoint,
            credentialsId,
            command,
            args,
            true)

        return new JsonSlurper().parseText(transportRequestId).REQUESTID
    }

    void uploadFileToTransportRequestSOLMAN(
        String changeId,
        String transportRequestId,
        String applicationId,
        String filePath,
        String endpoint,
        String credentialsId,
        String cmclientOpts = '') {

        def args = [
                '-cID', changeId,
                '-tID', transportRequestId,
                applicationId, "\"$filePath\""
            ]

        uploadFileToTransportRequest(
            BackendType.SOLMAN,
            '',
            [],
            endpoint,
            credentialsId,
            'upload-file-to-transport',
            args,
            cmclientOpts)
    }

    void uploadFileToTransportRequestCTS(
        String transportRequestId,
        String applicationId,
        String filePath,
        String endpoint,
        String credentialsId,
        String cmclientOpts = '') {

        def args = [
                '-tID', transportRequestId,
                "\"$filePath\""
            ]

        uploadFileToTransportRequest(
            BackendType.CTS,
            '',
            [],
            endpoint,
            credentialsId,
            'upload-file-to-transport',
            args,
            cmclientOpts)
    }

    void uploadFileToTransportRequestRFC(
        String dockerImage,
        List dockerOptions,
        String transportRequestId,
        String applicationId,
        String filePath,
        String endpoint,
        String credentialsId,
        String developmentInstance,
        String developmentClient,
        String applicationDescription,
        String abapPackage) {

        def args = [
                "--env ABAP_DEVELOPMENT_INSTANCE=${developmentInstance}",
                "--env ABAP_DEVELOPMENT_CLIENT=${developmentClient}",
                "--env ABAP_APPLICATION_NAME=${applicationId}",
                "--env ABAP_APPLICATION_DESC=${applicationDescription}",
                "--env ABAP_PACKAGE=${abapPackage}",
                "--env ZIP_FILE_URL=${filePath}",
            ]

            uploadFileToTransportRequest(
                BackendType.RFC,
                dockerImage,
                dockerOptions,
                endpoint,
                credentialsId,
                "cts uploadToABAP:${transportRequestId}",
                args,
                null)
    }

    private void uploadFileToTransportRequest(
        BackendType type,
        def dockerImage,
        List dockerOptions,
        def endpoint,
        def credentialsId,
        def command,
        List args,
        def cmclientOpts) {

        if(! type in [BackendType.SOLMAN, BackendType.CTS, BackendType.RFC]) {
            throw new IllegalArgumentException("Invalid backend type: ${type}")
        }

        int rc = executeWithCredentials(type,
                                        dockerImage,
                                        dockerOptions,
                                        endpoint,
                                        credentialsId,
                                        command,
                                        args,
                                        false,
                                        cmclientOpts) as int
        if(rc == 0) {
            return
        } else {
            throw new ChangeManagementException("Cannot upload file into transport request. Return code from cmclient: $rc.")
        }

    }

    def executeWithCredentials(BackendType type,
                               String dockerImage,
                               List dockerOptions,
                               String endpoint,
                               String credentialsId,
                               String command,
                               List args,
                               boolean returnStdout = false,
                               String clientOpts = '') {

       def script = this.script
       script.withCredentials([script.usernamePassword(
            credentialsId: credentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {

            Map shArgs = [:]

            if(returnStdout)
                shArgs.put('returnStdout', true)
            else
                shArgs.put('returnStatus', true)

            if(type == BackendType.RFC) {

                shArgs.script = command

                args = args.plus([
                    "--env ABAP_DEVELOPMENT_SERVER=${endpoint}",
                    "--env ABAP_DEVELOPMENT_USER=${script.username}",
                    "--env ABAP_DEVELOPMENT_PASSWORD=${script.password}"])

                dockerOptions = dockerOptions.plus(args)

                def result = 1

                script.dockerExecute(script: script,
                                     dockerImage: dockerImage,
                                     dockerOptions: dockerOptions ) {

                    result = script.sh(shArgs)

                }

                return result

            } else {

                def cmScript = getCMCommandLine(type, endpoint, script.username, script.password,
                    command, args,
                    clientOpts)

                shArgs.script = cmScript

                // user and password are masked by withCredentials
                script.echo """[INFO] Executing command line: "${cmScript}"."""
                return script.sh(shArgs)
            }
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

        int rc = executeWithCredentials(type, '', [], endpoint, credentialsId, cmd, args, false, clientOpts) as int
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
