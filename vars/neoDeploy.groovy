import com.sap.piper.Utils

import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger
import com.sap.piper.ConfigurationType
import com.sap.piper.tools.ToolDescriptor


def call(parameters = [:]) {

    def stepName = 'neoDeploy'

    Set parameterKeys = [
        'applicationName',
        'archivePath',
        'account',
        'deployAccount', //deprecated, replaced by parameter 'account'
        'deployHost', //deprecated, replaced by parameter 'host'
        'deployMode',
        'dockerEnvVars',
        'dockerImage',
        'dockerOptions',
        'host',
        'neoCredentialsId',
        'neoHome',
        'propertiesFile',
        'runtime',
        'runtimeVersion',
        'vmSize',
        'warAction'
        ]

    Set stepConfigurationKeys = [
        'account',
        'dockerEnvVars',
        'dockerImage',
        'dockerOptions',
        'host',
        'neoCredentialsId',
        'neoHome'
        ]

    handlePipelineStepErrors (stepName: stepName, stepParameters: parameters) {

        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        def utils = new Utils()

        prepareDefaultValues script: script

        final Map stepConfiguration = [:]

        // Backward compatibility: ensure old configuration is taken into account
        // The old configuration in not stage / step specific

        def defaultDeployHost = script.commonPipelineEnvironment.getConfigProperty('DEPLOY_HOST')
        if(defaultDeployHost) {
            echo "[WARNING][${stepName}] A deprecated configuration framework is used for configuring parameter 'DEPLOY_HOST'. This configuration framework will be removed in future versions."
            stepConfiguration.put('host', defaultDeployHost)
        }

        def defaultDeployAccount = script.commonPipelineEnvironment.getConfigProperty('CI_DEPLOY_ACCOUNT')
        if(defaultDeployAccount) {
            echo "[WARNING][${stepName}] A deprecated configuration framework is used for configuring parameter 'DEPLOY_ACCOUNT'. This configuration framekwork will be removed in future versions."
            stepConfiguration.put('account', defaultDeployAccount)
        }

        if(parameters.deployHost && !parameters.host) {
            echo "[WARNING][${stepName}] Deprecated parameter 'deployHost' is used. This will not work anymore in future versions. Use parameter 'host' instead."
            parameters.put('host', parameters.deployHost)
        }

        if(parameters.deployAccount && !parameters.account) {
            echo "[WARNING][${stepName}] Deprecated parameter 'deployAccount' is used. This will not work anymore in future versions. Use parameter 'account' instead."
            parameters.put('account', parameters.deployAccount)
        }

        def credId = script.commonPipelineEnvironment.getConfigProperty('neoCredentialsId')

        if(credId && !parameters.neoCredentialsId) {
            echo "[WARNING][${stepName}] Deprecated parameter 'neoCredentialsId' from old configuration framework is used. This will not work anymore in future versions."
            parameters.put('neoCredentialsId', credId)
        }

        // Backward compatibility end

        stepConfiguration.putAll(ConfigurationLoader.stepConfiguration(script, stepName))

        Map configuration = ConfigurationMerger.merge(parameters, parameterKeys,
                                                      stepConfiguration, stepConfigurationKeys,
                                                      ConfigurationLoader.defaultStepConfiguration(script, stepName))

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

        def neoVersions = ['neo-java-web': '3.39.10', 'neo-javaee6-wp': '2.132.6', 'neo-javaee7-wp': '1.21.13']
        def neo = new ToolDescriptor('SAP Cloud Platform Console Client', 'NEO_HOME', 'neoHome', '/tools/', 'neo.sh', neoVersions, 'version')
        def neoExecutable = neo.getToolExecutable(this, configuration)
        def neoDeployScript = """#!/bin/bash
                                 "${neoExecutable}" ${warAction} \
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

            def commonDeployParams =
                """--user '${username}' \
                   --password '${password}' \
                   --source "${archivePath}" \
                """
            dockerExecute(dockerImage: configuration.get('dockerImage'),
                          dockerEnvVars: configuration.get('dockerEnvVars'),
                          dockerOptions: configuration.get('dockerOptions')) {

                neo.verify(this, configuration)

                def java = new ToolDescriptor('Java', 'JAVA_HOME', '', '/bin/', 'java', '1.8.0', '-version 2>&1')
                java.verify(this, configuration)

                sh """${neoDeployScript} \
                      ${commonDeployParams}
                   """
            }
        }
    }
}
