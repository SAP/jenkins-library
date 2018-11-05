import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils

import com.sap.piper.tools.ToolDescriptor

import groovy.transform.Field

@Field String STEP_NAME = 'neoDeploy'
@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = [
    'account',
    'dockerEnvVars',
    'dockerImage',
    'dockerOptions',
    'host',
    'neoCredentialsId',
    'neoHome'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    'applicationName',
    'archivePath',
    'deployAccount', //deprecated, replaced by parameter 'account'
    'deployHost', //deprecated, replaced by parameter 'host'
    'deployMode',
    'propertiesFile',
    'runtime',
    'runtimeVersion',
    'vmSize',
    'warAction'
])

void call(parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this

        def utils = new Utils()

        prepareDefaultValues script: script

        final Map stepCompatibilityConfiguration = [:]

        // Backward compatibility: ensure old configuration is taken into account
        // The old configuration in not stage / step specific

        def defaultDeployHost = script.commonPipelineEnvironment.getConfigProperty('DEPLOY_HOST')
        if(defaultDeployHost) {
            echo "[WARNING][${STEP_NAME}] A deprecated configuration framework is used for configuring parameter 'DEPLOY_HOST'. This configuration framework will be removed in future versions."
            stepCompatibilityConfiguration.put('host', defaultDeployHost)
        }

        def defaultDeployAccount = script.commonPipelineEnvironment.getConfigProperty('CI_DEPLOY_ACCOUNT')
        if(defaultDeployAccount) {
            echo "[WARNING][${STEP_NAME}] A deprecated configuration framework is used for configuring parameter 'DEPLOY_ACCOUNT'. This configuration framekwork will be removed in future versions."
            stepCompatibilityConfiguration.put('account', defaultDeployAccount)
        }

        if(parameters.deployHost && !parameters.host) {
            echo "[WARNING][${STEP_NAME}] Deprecated parameter 'deployHost' is used. This will not work anymore in future versions. Use parameter 'host' instead."
            parameters.put('host', parameters.deployHost)
        }

        if(parameters.deployAccount && !parameters.account) {
            echo "[WARNING][${STEP_NAME}] Deprecated parameter 'deployAccount' is used. This will not work anymore in future versions. Use parameter 'account' instead."
            parameters.put('account', parameters.deployAccount)
        }

        def credId = script.commonPipelineEnvironment.getConfigProperty('neoCredentialsId')
        if(credId && !parameters.neoCredentialsId) {
            echo "[WARNING][${STEP_NAME}] Deprecated parameter 'neoCredentialsId' from old configuration framework is used. This will not work anymore in future versions."
            parameters.put('neoCredentialsId', credId)
        }
        // Backward compatibility end

        // load default & individual configuration
        Map configuration = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixin(stepCompatibilityConfiguration)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .addIfEmpty('archivePath', script.commonPipelineEnvironment.getMtarFilePath())
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        utils.pushToSWA([
            step: STEP_NAME,
            stepParam1: configuration.deployMode == 'mta'?'mta':'war', // ['mta', 'warParams', 'warPropertiesFile']
            stepParam2: configuration.warAction == 'rolling-update'?'blue-green':'standard', // ['deploy', 'deploy-mta', 'rolling-update']
            stepParam3: parameters?.script == null
        ], configuration)

        def archivePath = configuration.archivePath
        if(archivePath?.trim()) {
            if (!fileExists(archivePath)) {
                error "Archive cannot be found with parameter archivePath: '${archivePath}'."
            }
        } else {
            error "Archive path not configured (parameter \"archivePath\")."
        }

        def deployHost
        def deployAccount
        def credentialsId = configuration.get('neoCredentialsId')
        def deployMode = configuration.deployMode
        def warAction
        def propertiesFile
        def applicationName
        def runtime
        def runtimeVersion
        def vmSize

        def deployModes = ['mta', 'warParams', 'warPropertiesFile']
        if (! (deployMode in deployModes)) {
            throw new Exception("[neoDeploy] Invalid deployMode = '${deployMode}'. Valid 'deployMode' values are: ${deployModes}.")
        }

        if (deployMode in ['warPropertiesFile', 'warParams']) {
            warAction = utils.getMandatoryParameter(configuration, 'warAction')
            def warActions = ['deploy', 'rolling-update']
            if (! (warAction in warActions)) {
                throw new Exception("[neoDeploy] Invalid warAction = '${warAction}'. Valid 'warAction' values are: ${warActions}.")
            }
        } else if(deployMode == 'mta') {
            warAction = 'deploy-mta'
        }

        if (deployMode == 'warPropertiesFile') {
            propertiesFile = utils.getMandatoryParameter(configuration, 'propertiesFile')
            if (!fileExists(propertiesFile)){
                error "Properties file cannot be found with parameter propertiesFile: '${propertiesFile}'."
            }
        }

        if (deployMode == 'warParams') {
            applicationName = utils.getMandatoryParameter(configuration, 'applicationName')
            runtime = utils.getMandatoryParameter(configuration, 'runtime')
            runtimeVersion = utils.getMandatoryParameter(configuration, 'runtimeVersion')
            def vmSizes = ['lite', 'pro', 'prem', 'prem-plus']
            vmSize = configuration.vmSize
            if (! (vmSize in vmSizes)) {
                throw new Exception("[neoDeploy] Invalid vmSize = '${vmSize}'. Valid 'vmSize' values are: ${vmSizes}.")
            }
        }

        if (deployMode in ['mta','warParams']) {
            deployHost = utils.getMandatoryParameter(configuration, 'host')
            deployAccount = utils.getMandatoryParameter(configuration, 'account')
        }

        def neo = new ToolDescriptor('SAP Cloud Platform Console Client', 'NEO_HOME', 'neoHome', '/tools/', 'neo.sh', null, 'version')
        def neoExecutable = neo.getToolExecutable(this, configuration)
        def neoDeployScript = """#!/bin/bash
                                 "${neoExecutable}" ${warAction} \
                                 --source "${archivePath}" \
                              """

        if (deployMode in ['mta', 'warParams']) {
            neoDeployScript +=
                    """--host '${deployHost}' \
                    --account '${deployAccount}' \
                    """
        }

        if (deployMode == 'mta') {
            neoDeployScript += "--synchronous"
        }

        if (deployMode == 'warParams') {
            neoDeployScript +=
                    """--application '${applicationName}' \
                    --runtime '${runtime}' \
                    --runtime-version '${runtimeVersion}' \
                    --size '${vmSize}'"""
        }

        if (deployMode == 'warPropertiesFile') {
            neoDeployScript +=
                    """${propertiesFile}"""
        }

        withCredentials([usernamePassword(
            credentialsId: credentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {

            def credentials =
                """--user '${username}' \
                   --password '${password}' \
                """
            dockerExecute(dockerImage: configuration.get('dockerImage'),
                          dockerEnvVars: configuration.get('dockerEnvVars'),
                          dockerOptions: configuration.get('dockerOptions')) {

                neo.verify(this, configuration)

                def java = new ToolDescriptor('Java', 'JAVA_HOME', '', '/bin/', 'java', '1.8.0', '-version 2>&1')
                java.verify(this, configuration)

                sh """${neoDeployScript} \
                      ${credentials}
                   """
            }
        }
    }
}
