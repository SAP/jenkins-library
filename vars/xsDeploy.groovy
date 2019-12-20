import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.DefaultValueCache
import com.sap.piper.JenkinsUtils
import com.sap.piper.PiperGoUtils


import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils

import groovy.transform.Field

@Field String METADATA_FILE = 'metadata/xsDeploy.yaml'
@Field String PIPER_DEFAULTS = 'default_pipeline_environment.yml'
@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FOLDER = '.pipeline' // metadata file contains already the "metadata" folder level, hence we end up in a folder ".pipeline/metadata"
@Field String ADDITIONAL_CONFIGS_FOLDER='.pipeline/additionalConfigs'


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

        // hard to predict how these parameters looks like in its serialized form. Anyhow it is better
        // not to have these parameters forwarded somehow to the go layer.
        parameters.remove('juStabUtils')
        parameters.remove('piperGoUtils')
        parameters.remove('script')

        piperGoUtils.unstashPiperBin()

        //
        // Printing the piper-go version. Should not be done here, but somewhere during materializing
        // the piper binary. As long as we don't have it elsewhere we should keep it here.
        def piperGoVersion = sh(returnStdout: true, script: "./piper version")
        echo "PiperGoVersion: ${piperGoVersion}"

        //
        // since there is no valid config provided (... null) telemetry is disabled (same for other go releated steps at the moment).
        utils.pushToSWA([
            step: STEP_NAME,
        ], null)

        String configFiles = prepareConfigurations([PIPER_DEFAULTS].plus(script.commonPipelineEnvironment.getCustomDefaults()), ADDITIONAL_CONFIGS_FOLDER)

        writeFile(file: "${METADATA_FOLDER}/${METADATA_FILE}", text: libraryResource(METADATA_FILE))

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(parameters)}",
        ]) {

            //
            // context config gives us e.g. the docker image name. --> How does this work for customer maintained images?
            // There is a name provided in the metadata file. But we do not provide a docker image for that.
            // The user has to build that for her/his own. How do we expect to configure this?

            String projectConfigScript = "./piper getConfig --stepMetadata '${METADATA_FOLDER}/${METADATA_FILE}' --defaultConfig ${configFiles}"
            String contextConfigScript = projectConfigScript + " --contextConfig"
            Map projectConfig = readJSON (text: sh(returnStdout: true, script: projectConfigScript))
            Map contextConfig = readJSON (text: sh(returnStdout: true, script: contextConfigScript))

            Action action = projectConfig.action
            DeployMode mode = projectConfig.mode

            if(parameters.verbose) {
                echo "[INFO] ContextConfig: ${contextConfig}"
                echo "[INFO] ProjectConfig: ${projectConfig}"
            }

            def operationId
            if(mode == DeployMode.BG_DEPLOY && action != Action.NONE) {
                operationId = script.commonPipelineEnvironment.xsDeploymentId
                if (! operationId) {
                    throw new IllegalArgumentException('No operationId provided. Was there a deployment before?')
                }
            }

            def xsDeployStdout

            lock(getLockIdentifier(projectConfig)) {

                withCredentials([usernamePassword(
                        credentialsId: contextConfig.credentialsId,
                        passwordVariable: 'PASSWORD',
                        usernameVariable: 'USERNAME')]) {

                    dockerExecute([script: this].plus([dockerImage: contextConfig.dockerImage, dockerPullImage: false])) {
                        xsDeployStdout = sh returnStdout: true, script: """#!/bin/bash
                        ./piper xsDeploy --defaultConfig ${configFiles} --user \${USERNAME} --password \${PASSWORD} ${operationId ? "--operationId " + operationId : "" }
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

/*
 * The returned string can be used directly in the command line for retrieving the configuration via go
 */
String prepareConfigurations(List configs, String configCacheFolder) {

    for(def customDefault : configs) {
        writeFile(file: "${ADDITIONAL_CONFIGS_FOLDER}/${customDefault}", text: libraryResource(customDefault))
    }
    joinAndQuote(configs.reverse(), configCacheFolder)
}

/*
 * prefix is supposed to be provided without trailing slash
 */
String joinAndQuote(List l, String prefix = '') {
    _l = []

    if(prefix == null) {
        prefix = ''
    }
    if(prefix.endsWith('/') || prefix.endsWith('\\'))
        throw new IllegalArgumentException("Provide prefix (${prefix}) without trailing slash")

    for(def e : l) {
         def _e = ''
         if(prefix.length() > 0) {
             _e += prefix
             _e += '/'
         }
         _e += e 
        _l << '"' + _e + '"'
    }
    _l.join(' ')
}
