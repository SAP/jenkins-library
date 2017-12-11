import com.sap.piper.Utils


def call(parameters = [:]) {

    handlePipelineStepErrors (stepName: 'neoDeploy', stepParameters: parameters) {

        def utils = new Utils()

        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        def archivePath = utils.getMandatoryParameter(parameters, 'archivePath', null)
        if (!fileExists(archivePath)){
            error "Archive cannot be found with parameter archivePath: '${archivePath}'."
        }

        def deployMode = utils.getMandatoryParameter(parameters, 'deployMode', 'mta')

        if (deployMode != 'mta' && deployMode != 'warParams' && deployMode != 'warPropertiesFile') {
            throw new Exception("[neoDeploy] Invalid deployMode = '${deployMode}'. Valid 'deployMode' values are: 'mta', 'warParams' and 'warPropertiesFile'")
        }

        def propertiesFile
        def warAction
        if (deployMode == 'warPropertiesFile' || deployMode == 'warParams') {
            warAction = utils.getMandatoryParameter(parameters, 'warAction', 'deploy')
            if (warAction != 'deploy' && warAction != 'rolling-update') {
                throw new Exception("[neoDeploy] Invalid warAction = '${warAction}'. Valid 'warAction' values are: 'deploy' and 'rolling-update'.")
            }
        }
        if (deployMode == 'warPropertiesFile') {
            propertiesFile = utils.getMandatoryParameter(parameters, 'propertiesFile', null)
            if (!fileExists(propertiesFile)){
                error "Properties file cannot be found with parameter propertiesFile: '${propertiesFile}'."
            }
        }

        def applicationName
        def runtime
        def runtimeVersion
        def vmSize
        if (deployMode == 'warParams') {
            applicationName = utils.getMandatoryParameter(parameters, 'applicationName', null)
            runtime = utils.getMandatoryParameter(parameters, 'runtime', null)
            runtimeVersion = utils.getMandatoryParameter(parameters, 'runtimeVersion', null)
            vmSize = utils.getMandatoryParameter(parameters, 'vmSize', 'lite')
            if (vmSize != 'lite' && vmSize !='pro' && vmSize != 'prem' && vmSize != 'prem-plus') {
                throw new Exception("[neoDeploy] Invalid vmSize = '${vmSize}'. Valid 'vmSize' values are: 'lite', 'pro', 'prem' and 'prem-plus'.")
            }
        }

        def defaultDeployHost = script.commonPipelineEnvironment.getConfigProperty('DEPLOY_HOST')
        def defaultDeployAccount = script.commonPipelineEnvironment.getConfigProperty('CI_DEPLOY_ACCOUNT')
        def defaultCredentialsId = script.commonPipelineEnvironment.getConfigProperty('neoCredentialsId')
        if (defaultCredentialsId == null) {
            defaultCredentialsId = 'CI_CREDENTIALS_ID'
        }

        def deployHost
        def deployAccount

            if (deployMode.equals('mta') || deployMode.equals('warParams')) {
            deployHost = utils.getMandatoryParameter(parameters, 'deployHost', defaultDeployHost)
            deployAccount = utils.getMandatoryParameter(parameters, 'deployAccount', defaultDeployAccount)
        }

        def credentialsId = parameters.get('neoCredentialsId', defaultCredentialsId)

        def neoExecutable = getNeoExecutable(parameters)

        withCredentials([usernamePassword(
            credentialsId: credentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {

            def commonDeployParams =
                """--user '${username}' \
                   --password '${password}' \
                   --source "${archivePath}" \
                """

            if (deployMode == 'mta') {
                sh """#!/bin/bash
                      "${neoExecutable}" deploy-mta \
                      ${commonDeployParams} \
                      --host '${deployHost}' \
                      --account '${deployAccount}' \
                      --synchronous
                   """
            }

            if (deployMode == 'warParams') {
                sh """#!/bin/bash
                      "${neoExecutable}" ${warAction} \
                      ${commonDeployParams} \
                      --host '${deployHost}' \
                      --account '${deployAccount}' \
                      --application '${applicationName}' \
                      --runtime '${runtime}' \
                      --runtime-version '${runtimeVersion}' \
                      --size '${vmSize}'
                   """
            }

            if (deployMode == 'warPropertiesFile') {
                sh """#!/bin/bash
                      "${neoExecutable}" ${warAction} \
                      ${commonDeployParams} \
                      ${propertiesFile}
                   """
            }
        }
    }
}

private getNeoExecutable(parameters) {

    def neoExecutable = 'neo' // default, if nothing below applies maybe it is the path.

    if (parameters?.neoHome) {
        neoExecutable = "${parameters.neoHome}/tools/neo.sh"
        echo "[neoDeploy] Neo executable \"${neoExecutable}\" retrieved from parameters."
        return neoExecutable
    }

    if (env?.NEO_HOME) {
        neoExecutable = "${env.NEO_HOME}/tools/neo.sh"
        echo "[neoDeploy] Neo executable \"${neoExecutable}\" retrieved from environment."
        return neoExecutable
    }

    echo "Using Neo executable from PATH."
    return neoExecutable
}
