import com.sap.piper.JenkinsUtils

import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.BashUtils
import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils

import groovy.transform.Field

import hudson.AbortException

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

@Field Set STEP_CONFIG_KEYS = [
    'action',
    'apiUrl',
    'credentialsId',
    'deploymentId',
    'deployIdLogPattern',
    'deployOpts',
    /** A map containing properties forwarded to dockerExecute. For more details see [here][dockerExecute] */
    'docker',
    'loginOpts',
    'mode',
    'mtaPath',
    'org',
    'space',
    'xsSessionFile',
]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

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

/**
  * Performs an XS deployment
  *
  * In case of blue-green deployments the step is called for the deployment in the narrower sense
  * and later again or resuming or aborting. In this case both calls needs to be performed from the
  * same directory.
  */
@GenerateDocumentation
void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def utils = parameters.juStabUtils ?: new Utils()

        final script = checkScript(this, parameters) ?: this

        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .addIfEmpty('mtaPath', script.commonPipelineEnvironment.getMtarFilePath())
            .addIfEmpty('deploymentId', script.commonPipelineEnvironment.xsDeploymentId)
            .mixin(parameters, PARAMETER_KEYS)

        Map config = configHelper.use()

        DeployMode mode = config.mode

        if(mode == DeployMode.NONE) {
            echo "Deployment skipped intentionally. Deploy mode '${mode.toString()}'."
            return
        }

        Action action = config.action

        if(mode == DeployMode.DEPLOY && action != Action.NONE) {
            error "Cannot perform action '${action.toString()}' in mode '${mode.toString()}'. Only action '${Action.NONE.toString()}' is allowed."
        }

        boolean performLogin  = ((mode == DeployMode.DEPLOY) || (mode == DeployMode.BG_DEPLOY && !(action in [Action.RESUME, Action.ABORT])))
        boolean performLogout = ((mode == DeployMode.DEPLOY) || (mode == DeployMode.BG_DEPLOY && action != Action.NONE))

        boolean sessionExists = fileExists file: config.xsSessionFile

        if( (! performLogin) && (! sessionExists) ) {
            error 'For the current configuration an already existing session is required. But there is no already existing session.'
        }

        configHelper
            .collectValidationFailures()
            /**
              * Used for finalizing the blue-green deployment.
              * @possibleValues RESUME, ABORT, RETRY
              */
            .withMandatoryProperty('action')
            /** The file name of the file representing the sesssion after `xs login`. Should not be changed normally. */
            .withMandatoryProperty('xsSessionFile')
            /** Regex pattern for retrieving the ID of the deployment. */
            .withMandatoryProperty('deployIdLogPattern')
            /**
              * Controls if there is a standard deployment or a blue green deployment
              * @possibleValues DEPLOY, BG_DEPLOY
              */
            .withMandatoryProperty('mode')
            /** The endpoint */
            .withMandatoryProperty('apiUrl')
            /** The organization */
            .withMandatoryProperty('org')
            /** The space */
            .withMandatoryProperty('space')
            /** Additional options appended to the login command. Only needed for sophisticated cases.
              * When provided it is the duty of the provider to ensure proper quoting / escaping.
              */
            .withMandatoryProperty('loginOpts')
            /** Additional options appended to the deploy command. Only needed for sophisticated cases.
              * When provided it is the duty of the provider to ensure proper quoting / escaping.
              */
            .withMandatoryProperty('deployOpts')
            /** The credentialsId */
            .withMandatoryProperty('credentialsId')
            /** The path to the deployable. If not provided explicitly it is retrieved from the common pipeline environment
              * (Parameter `mtarFilePath`).
              */
            .withMandatoryProperty('mtaPath', null, {action == Action.NONE})
            .withMandatoryProperty('deploymentId',
                'No deployment id provided, neither via parameters nor via common pipeline environment. Was there a deployment before?',
                {action in [Action.RESUME, Action.ABORT, Action.RETRY]})
            .use()

        utils.pushToSWA([
            step: STEP_NAME,
        ], config)

        if(performLogin) {
            login(script, config)
        }

        def failures = []

        if(action in [Action.RESUME, Action.ABORT, Action.RETRY]) {

            complete(script, mode, action, config, failures)

        } else {

            deploy(script, mode, config, failures)
        }

        if (performLogout || failures) {
            logout(script, config, failures)

        } else {
            echo "Skipping logout in order to be able to resume or abort later."
        }

        if(failures) {
            error "Failed command(s): ${failures}. Check earlier log for details."
        }
    }
}

void login(Script script, Map config) {

    withCredentials([usernamePassword(
        credentialsId: config.credentialsId,
        passwordVariable: 'password',
        usernameVariable: 'username'
    )]) {

        def returnCode = executeXSCommand([script: script].plus(config.docker),
        [
            "xs login -a ${config.apiUrl} -u ${username} -p ${BashUtils.quoteAndEscape(password)} -o ${config.org} -s ${config.space} ${config.loginOpts}",
            'RC=$?',
            "[ \$RC == 0 ]  && cp \"\${HOME}/${config.xsSessionFile}\" .",
            'exit $RC'
        ])

        if(returnCode != 0)
            error "xs login failed."
    }

    boolean existsXsSessionFileAfterLogin = fileExists file: config.xsSessionFile
    if(! existsXsSessionFileAfterLogin)
        error "Session file ${config.xsSessionFile} not found in current working directory after login."
}

void deploy(Script script, DeployMode mode, Map config, def failures) {

    def deploymentLog

    try {
        lock(getLockIdentifier(config)) {
            deploymentLog = executeXSCommand([script: script].plus(config.docker),
            [
                "cp ${config.xsSessionFile} \${HOME}",
                "xs ${mode.toString()} '${config.mtaPath}' -f ${config.deployOpts}"
            ], true)
        }

        echo "Deploy log: ${deploymentLog}"

    } catch(AbortException e) {
        echo "deployment failed. Message: ${e.getMessage()}, Log: ${deploymentLog}}"
        failures << "xs ${mode.toString()}"
    }

    if(mode == DeployMode.BG_DEPLOY) {

        if(! failures.isEmpty()) {

            echo "Retrieval of deploymentId skipped since prior deployment was not successfull."

        } else {

            for (def logLine : deploymentLog.readLines()) {
                def matcher = logLine =~ config.deployIdLogPattern
                if(matcher.find()) {
                    script.commonPipelineEnvironment.xsDeploymentId = matcher[0][1]
                    echo "DeploymentId: ${script.commonPipelineEnvironment.xsDeploymentId}."
                    break
                }
            }
            if(script.commonPipelineEnvironment.xsDeploymentId == null) {
                failures << "Cannot lookup deploymentId. Search pattern was: '${config.deployIdLogPattern}'."
            }
        }
    }
}

void complete(Script script, DeployMode mode, Action action, Map config, def failures) {

    if(mode != DeployMode.BG_DEPLOY)
        error "Action '${action.toString()}' can only be performed for mode '${DeployMode.BG_DEPLOY.toString()}'. Current mode is: '${mode.toString()}'."

    def returnCode = 1

    lock(getLockIdentifier(config)) {
        returnCode = executeXSCommand([script: script].plus(config.docker),
        [
            "cp ${config.xsSessionFile} \${HOME}",
            "xs ${mode.toString()} -i ${config.deploymentId} -a ${action.toString()}"
        ])
    }

    if(returnCode != 0) {
        echo "${mode.toString()} with action '${action.toString()}' failed with return code ${returnCode}."
        failures << "xs ${mode.toString()} -a ${action.toString()}"
    }
}

void logout(Script script, Map config, def failures) {

    def returnCode = executeXSCommand([script: script].plus(config.docker),
    [
        "cp ${config.xsSessionFile} \${HOME}",
        'xs logout'
    ])

    if(returnCode != 0) {
        failures << 'xs logout'
    }

    sh "XSCONFIG=${config.xsSessionFile}; [ -f \${XSCONFIG} ] && rm \${XSCONFIG}"
}

String getLockIdentifier(Map config) {
    "$STEP_NAME:${config.apiUrl}:${config.org}:${config.space}"
}

def executeXSCommand(Map dockerOptions, List commands, boolean returnStdout = false) {

    def r

    dockerExecute(dockerOptions) {

        // in case there are credentials contained in the commands we assume
        // the call is properly wrapped by withCredentials(./.)
        echo "Executing: '${commands}'."

        List prelude = [
            '#!/bin/bash'
        ]

        List script = (prelude + commands)

        params = [
            script: script.join('\n')
        ]

        if(returnStdout) {
            params << [ returnStdout: true ]
        } else {
            params << [ returnStatus: true ]
        }

        r = sh params

        if( (! returnStdout ) && r != 0) {

            try {
                echo "xs logs:"

                sh 'LOG_FOLDER=${HOME}/.xs_logs; [ -d ${LOG_FOLDER} ]  && cat ${LOG_FOLDER}/*'

            } catch(Exception e) {

                echo "Cannot provide xs logs: ${e.getMessage()}."
            }

            echo "Executing of commands '${commands}' failed. Check earlier logs for details."
        }
    }
    r
}
