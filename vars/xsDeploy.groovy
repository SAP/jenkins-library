import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.JenkinsUtils
import com.sap.piper.PiperGoUtils


import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils

import groovy.transform.Field

@Field String METADATA_FILE = 'metadata/xsDeploy.yaml'
@Field String STEP_NAME = getClass().getName()


enum DeployMode {
    DEPLOY,
    BG_DEPLOY,
    NONE

    String toString() {
        name().toLowerCase(Locale.ENGLISH).replaceAll('_', '-')
    }
}

enum Action {
    RESUME,
    ABORT,
    RETRY,
    NONE

    String toString() {
        name().toLowerCase(Locale.ENGLISH)
    }
}

void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        final script = checkScript(this, parameters) ?: null

        if(! script) {
            error "Reference to surrounding pipeline script not provided (script: this)."
        }

        def utils = parameters.juStabUtils ?: new Utils()
        def piperGoUtils = parameters.piperGoUtils ?: new PiperGoUtils(utils)

        //
        // The parameters map in provided from outside. That map might be used elsewhere in the pipeline
        // hence we should not modify it here. So we create a new map based on the parameters map.
        parameters = [:] << parameters

        // hard to predict how these two parameters looks like in its serialized form. Anyhow it is better
        // not to have these parameters forwarded somehow to the go layer.
        parameters.remove('juStabUtils')
        parameters.remove('piperGoUtils')
        parameters.remove('script')

        //
        // For now - since the xsDeploy step is not merged and covered by a release - we stash
        // a locally built version of the piper-go binary in the pipeline script (Jenkinsfile) with
        // stash name "piper-bin". That stash is used inside method "unstashPiperBin".
        piperGoUtils.unstashPiperBin()

        //
        // Printing the piper-go version. Should not be done here, but somewhere during materializing
        // the piper binary.
        def piperGoVersion = sh(returnStdout: true, script: "./piper version")
        echo "PiperGoVersion: ${piperGoVersion}"

        //
        // since there is no valid config provided (... null) telemetry is disabled.
        utils.pushToSWA([
            step: STEP_NAME,
        ], null)

        writeFile(file: METADATA_FILE, text: libraryResource(METADATA_FILE))


        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(parameters)}",
        ]) {

            //
            // context config gives us e.g. the docker image name. --> How does this work for customer maintained images?
            // There is a name provided in the metadata file. But we do not provide a docker image for that.
            // The user has to build that for her/his own. How do we expect to configure this?
            Map contextConfig = readJSON (text: sh(returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '${METADATA_FILE}'"))

            Map projectConfig = readJSON (text: sh(returnStdout: true, script: "./piper ${parameters.verbose ? '--verbose' :''} getConfig --stepMetadata '${METADATA_FILE}'"))

            if(parameters.verbose) {
                echo "[INFO] Context-Config: ${contextConfig}"
                echo "[INFO] Project-Config: ${projectConfig}"
            }

            Action action = projectConfig.action
            DeployMode mode = projectConfig.mode

            // That config map here is only used in the groovy layer. Nothing is handed over to go.
            Map config = contextConfig <<
                [
                    apiUrl: projectConfig.apiUrl, // required on groovy level for acquire the lock
                    org: projectConfig.org,       // required on groovy level for acquire the lock
                    space: projectConfig.space,   // required on groovy level for acquire the lock
                    docker: [
                        dockerImage: contextConfig.dockerImage,
                        dockerPullImage: false    // dockerPullImage apparently not provided by context config.
                    ]
                ]

            if(parameters.verbose) {
                echo "[INFO] Merged-Config: ${config}"
            }

            def operationId
            if(mode == DeployMode.BG_DEPLOY && action != Action.NONE) {
                operationId = script.commonPipelineEnvironment.xsDeploymentId
                if (! operationId) {
                    throw new IllegalArgumentException('No operationId provided. Was there a deployment before?')
                }
            }

            def xsDeployStdout

            lock(getLockIdentifier(config)) {

                withCredentials([usernamePassword(
                        credentialsId: config.credentialsId,
                        passwordVariable: 'PASSWORD',
                        usernameVariable: 'USERNAME')]) {

                    dockerExecute([script: this].plus(config.docker)) {
                        xsDeployStdout = sh returnStdout: true, script: """#!/bin/bash
                        ./piper ${parameters.verbose ? '--verbose' : ''} xsDeploy --user \${USERNAME} --password \${PASSWORD} ${operationId ? "--operationId " + operationId : "" }
                        """
                    }

                }
            }

            if(mode == DeployMode.BG_DEPLOY && action == Action.NONE) {
                script.commonPipelineEnvironment.xsDeploymentId = readJSON(text: xsDeployStdout).operationId
                if (!script.commonPipelineEnvironment.xsDeploymentId) {
                    error "No Operation id returned from xs deploy step. This is required for mode '${mode}' and action '${action}'."
                }
                echo "[INFO] OperationId for subsequent resume or abort: '${script.commonPipelineEnvironment.xsDeploymentId}'."
            }
        }
    }
}

String getLockIdentifier(Map config) {
    "$STEP_NAME:${config.apiUrl}:${config.org}:${config.space}"
}
