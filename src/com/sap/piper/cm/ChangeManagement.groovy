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

    boolean isChangeInDevelopment(Map docker, String changeId, String endpoint, String credentialsId, String clientOpts = '') {
        int rc = executeWithCredentials(BackendType.SOLMAN, docker, endpoint, credentialsId, 'is-change-in-development', ['-cID', "'${changeId}'", '--return-code'],
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

    String createTransportRequestCTS(Map docker, String transportType, String targetSystemId, String description, String endpoint, String credentialsId, String clientOpts = '') {
        try {
            def transportRequest = executeWithCredentials(BackendType.CTS, docker, endpoint, credentialsId, 'create-transport',
                    ['-tt', transportType, '-ts', targetSystemId, '-d', "\"${description}\""],
                    true,
                    clientOpts)
            return (transportRequest as String)?.trim()
        }catch(AbortException e) {
            throw new ChangeManagementException("Cannot create a transport request. $e.message.")
        }
    }

    String createTransportRequestSOLMAN(Map docker, String changeId, String developmentSystemId, String endpoint, String credentialsId, String clientOpts = '') {

        try {
            def transportRequest = executeWithCredentials(BackendType.SOLMAN, docker, endpoint, credentialsId, 'create-transport', ['-cID', changeId, '-dID', developmentSystemId],
                true,
                clientOpts)
            return (transportRequest as String)?.trim()
        }catch(AbortException e) {
            throw new ChangeManagementException("Cannot create a transport request for change id '$changeId'. $e.message.")
        }
    }

    String createTransportRequestRFC(
        Map docker,
        String endpoint,
        String developmentInstance,
        String developmentClient,
        String credentialsId,
        String description,
        boolean verbose) {

        def command = 'cts createTransportRequest'
        def args = [
            TRANSPORT_DESCRIPTION: description,
            ABAP_DEVELOPMENT_INSTANCE: developmentInstance,
            ABAP_DEVELOPMENT_CLIENT: developmentClient,
            VERBOSE: verbose,
        ]

        try {

            def transportRequestId = executeWithCredentials(
                BackendType.RFC,
                docker,
                endpoint,
                credentialsId,
                command,
                args,
                true)

            return new JsonSlurper().parseText(transportRequestId).REQUESTID

        } catch(AbortException ex) {
            throw new ChangeManagementException(
                "Cannot create transport request: ${ex.getMessage()}", ex)
        }
    }

    void uploadFileToTransportRequestSOLMAN(
        Map docker,
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

        int rc = executeWithCredentials(
            BackendType.SOLMAN,
            docker,
            endpoint,
            credentialsId,
            'upload-file-to-transport',
            args,
            false,
            cmclientOpts) as int

        if(rc != 0) {
            throw new ChangeManagementException(
                "Cannot upload file into transport request. Return code from cm client: $rc.")
        }
    }

    void uploadFileToTransportRequestCTS(
        Map docker,
        String transportRequestId,
        String endpoint,
        String client,
        String applicationName,
        String description,
        String abapPackage, // "package" would be better, but this is a keyword
        String osDeployUser,
        def deployToolDependencies,
        def npmInstallOpts,
        String deployConfigFile,
        String credentialsId) {

        def script = this.script

        def desc = description ?: 'Deployed with Piper based on SAP Fiori tools'

        /*
            Create the config file

            There are currently pull request for making more (all?) parameters configurable as
            command line parameters. With that there would be no need for creating this file here.
            Parameters could be provided via command line in case they are defined. If not, as a fallback,
            parameters define in the config file can be used.
            Currently not clear if also the credentials will be made available as command line parameters.
            If not we have to discuss how to deal wih the credentials. One option would be to parse an existing
            config file and either to inject the environment variables according what is defined there, or to update
            the config wrt credentials (credentials can be configured in the config file to be retrieved from
            environment variables).
            Also not clear if the config file is fully optional (can be omitted). If not and no config file is present
            we can either raise an exception (and ask the project owner for adding a config file) or add a dummy config file.
        */
        def deployConfig =  ("""|specVersion: '1.0'
                                |metadata:
                                |  name: ${applicationName}
                                |type: application
                                |builder:
                                |  customTasks:
                                |  - name: deploy-to-abap
                                |    afterTask: replaceVersion
                                |    configuration:
                                |      target:
                                |        client: ''
                                |        auth: basic
                                |      credentials:
                                |        username: env:ABAP_USER
                                |        password: env:ABAP_PASSWORD
                                |      app:
                                |        name: ''
                                |        description: ${desc}
                                |        package: ''
                                |      exclude:
                                |      - .*\\.test.js
                                |      - internal.md
                                |""" as CharSequence).stripMargin()

        script.writeFile file: deployConfigFile, text: deployConfig, encoding: 'UTF-8'

        if (deployToolDependencies in List) {
            deployToolDependencies = deployToolDependencies.join(' ')
        }

        if (npmInstallOpts in List) {
            npmInstallOpts = npmInstallOpts.join(' ')
        }

        deployToolDependencies = deployToolDependencies.trim()

        /*
            In case the configuration has been adjusted so that no deployToolDependencies are provided
            we assume an image is used which contains already all dependencies.
            In this case we don't invoke npm install and we run the image with the standard user
            already, since there is no need for being root. Hence we don't need to switch user also
            in the script.
         */
        boolean noInstall = deployToolDependencies.isEmpty()

        Iterable cmd = ['#!/bin/bash -e']

        if (! noInstall) {
            cmd << "npm install --global ${npmInstallOpts} ${deployToolDependencies}"
            cmd << "su ${osDeployUser}"
        } else {
            script.echo "[INFO] no deploy dependencies provided. Skipping npm install call. Assuning docker image '${docker?.image}' contains already the dependencies for performing the deployment."
        }

        Iterable params = []

        if (deployConfigFile) {
            /*
                not sure if we will manage to that optional
                the plan is to make all parameters configurable via command line parameters. In that case it would
                not be required anymore to have a config file at all. Open questions: credentials, excludes.
            */
            params += ['-c', "\"" + deployConfigFile + "\""]
        }
        if (transportRequestId) {
            params += ['-t', transportRequestId]
        }

        if (endpoint) {
            params += ['-u', endpoint]
        }

        params += ['-p', abapPackage]

        params += ['-n' , applicationName]

        params += ['-l', client]

        params += ['-f'] // failfast --> provide return code != 0 in case of any failure

        // more parameters can be added when they are recognized by the fiori toolset, e.g. abap package.


        params += ['-y'] // autoconfirm

        def fioriDeployCmd = "fiori deploy ${params.join(' ')}"
        script.echo "Executing deploy command: '${fioriDeployCmd}'"
        cmd << fioriDeployCmd

        script.withCredentials([script.usernamePassword(
            credentialsId: credentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {

            /*
                The config file is configured to read the credentials from the environment,
                see above in the config file template.
                After installing the deploy toolset we switch the user. Since we do not su with option '-l' the
                environment variables are preserved.
                Not clear of the credentials will also be available as command line parameters.
            */
            def dockerEnvVars = docker.envVars ?: [:] + [ABAP_USER: script.username, ABAP_PASSWORD: script.password]

            def dockerOptions = docker.options ?: []
            if (!noInstall) {
                // when we install globally we need to be root, after preparing that we can su node` in the bash script.
                // in case there is already a u provided the latest (... what we set here wins).
                dockerOptions += ['-u', '0']
            }

            script.dockerExecute(
                script: script,
                dockerImage: docker.image,
                dockerOptions: dockerOptions,
                dockerEnvVars: dockerEnvVars,
                dockerPullImage: docker.pullImage) {

                script.sh script: cmd.join('\n')
            }
        }
    }

    void uploadFileToTransportRequestRFC(
        Map docker,
        String transportRequestId,
        String applicationName,
        String filePath,
        String endpoint,
        String credentialsId,
        String developmentInstance,
        String developmentClient,
        String applicationDescription,
        String abapPackage,
        String codePage,
        boolean acceptUnixStyleEndOfLine,
        boolean failOnWarning,
        boolean verbose) {

        def args = [
            ABAP_DEVELOPMENT_INSTANCE: developmentInstance,
            ABAP_DEVELOPMENT_CLIENT: developmentClient,
            ABAP_APPLICATION_NAME: applicationName,
            ABAP_APPLICATION_DESC: applicationDescription,
            ABAP_PACKAGE: abapPackage,
            ZIP_FILE_URL: filePath,
            CODE_PAGE: codePage,
            ABAP_ACCEPT_UNIX_STYLE_EOL: acceptUnixStyleEndOfLine ? 'X' : '-',
            FAIL_UPLOAD_ON_WARNING: Boolean.toString(failOnWarning),
            VERBOSE: Boolean.toString(verbose),
        ]

        int rc = executeWithCredentials(
            BackendType.RFC,
            docker,
            endpoint,
            credentialsId,
            "cts uploadToABAP:${transportRequestId}",
            args,
            false) as int

        if(rc != 0) {
            throw new ChangeManagementException(
                "Cannot upload file into transport request. Return code from rfc client: $rc.")
        }
    }

    def executeWithCredentials(
        BackendType type,
        Map docker,
        String endpoint,
        String credentialsId,
        String command,
        def args,
        boolean returnStdout = false,
        String clientOpts = '') {

        def script = this.script

        docker = docker ?: [:]

        script.withCredentials([script.usernamePassword(
            credentialsId: credentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {

            Map shArgs = [:]

            if(returnStdout)
                shArgs.put('returnStdout', true)
            else
                shArgs.put('returnStatus', true)

            Map dockerEnvVars = docker.envVars ?: [:]

            def result = 1

            switch(type) {

                case BackendType.RFC:

                    if(! (args in Map)) {
                        throw new IllegalArgumentException("args expected as Map for backend types ${[BackendType.RFC]}")
                    }

                    shArgs.script = command

                    args = args.plus([
                        ABAP_DEVELOPMENT_SERVER: endpoint,
                        ABAP_DEVELOPMENT_USER: script.username,
                        ABAP_DEVELOPMENT_PASSWORD: script.password,
                    ])

                    dockerEnvVars += args

                    break

                case BackendType.SOLMAN:
                case BackendType.CTS:

                    if(! (args in Collection))
                        throw new IllegalArgumentException("args expected as Collection for backend types ${[BackendType.SOLMAN, BackendType.CTS]}")

                    shArgs.script = getCMCommandLine(type, endpoint, script.username, script.password,
                        command, args,
                        clientOpts)

                    break
            }

        // user and password are masked by withCredentials
        script.echo """[INFO] Executing command line: "${shArgs.script}"."""

                script.dockerExecute(
                    script: script,
                    dockerImage: docker.image,
                    dockerOptions: docker.options,
                    dockerEnvVars: dockerEnvVars,
                    dockerPullImage: docker.pullImage) {

                    result = script.sh(shArgs)

                    }

            return result
        }
    }

    void releaseTransportRequestSOLMAN(
        Map docker,
        String changeId,
        String transportRequestId,
        String endpoint,
        String credentialsId,
        String clientOpts = '') {

        def cmd = 'release-transport'
        def args = [
            '-cID',
            changeId,
            '-tID',
            transportRequestId,
        ]

        int rc = executeWithCredentials(
            BackendType.SOLMAN,
            docker,
            endpoint,
            credentialsId,
            cmd,
            args,
            false,
            clientOpts) as int

        if(rc != 0) {
            throw new ChangeManagementException("Cannot release Transport Request '$transportRequestId'. Return code from cmclient: $rc.")
        }
    }

    void releaseTransportRequestCTS(
        Map docker,
        String transportRequestId,
        String endpoint,
        String credentialsId,
        String clientOpts = '') {

        def cmd = 'export-transport'
        def args = [
            '-tID',
            transportRequestId,
        ]

        int rc = executeWithCredentials(
            BackendType.CTS,
            docker,
            endpoint,
            credentialsId,
            cmd,
            args,
            false) as int

        if(rc != 0) {
            throw new ChangeManagementException("Cannot release Transport Request '$transportRequestId'. Return code from cmclient: $rc.")
        }
    }

    void releaseTransportRequestRFC(
        Map docker,
        String transportRequestId,
        String endpoint,
        String developmentInstance,
        String developmentClient,
        String credentialsId,
        boolean verbose) {

        def cmd = "cts releaseTransport:${transportRequestId}"
        def args = [
            ABAP_DEVELOPMENT_INSTANCE: developmentInstance,
            ABAP_DEVELOPMENT_CLIENT: developmentClient,
            VERBOSE: verbose,
        ]

        int rc = executeWithCredentials(
            BackendType.RFC,
            docker,
            endpoint,
            credentialsId,
            cmd,
            args,
            false) as int

        if(rc != 0) {
            throw new ChangeManagementException("Cannot release Transport Request '$transportRequestId'. Return code from rfcclient: $rc.")
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
            cmCommandLine += """
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
