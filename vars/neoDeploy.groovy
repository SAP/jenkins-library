import com.sap.piper.Utils


def call(parameters = [:]) {

    handlePipelineStepErrors (stepName: 'neoDeploy', stepParameters: parameters) {

        def utils = new Utils()
        def script = parameters.script
        if (script == null){
            script = [commonPipelineEnvironment: commonPipelineEnvironment]
        }

        def archivePath = new File(utils.getMandatoryParameter(parameters, 'archivePath', null))
        if (!archivePath.isAbsolute()) {
            archivePath = new File(pwd(), archivePath.getPath())
        }
        if (!archivePath.exists()){
            error "Archive cannot be found with parameter archivePath: '${archivePath}'."
        }

        def deployMode = utils.getMandatoryParameter(parameters, 'deployMode', 'MTA')

        if (deployMode != 'MTA' && deployMode != 'WAR_PARAMS' && deployMode != 'WAR_PROPERTIESFILE') {
            throw new IllegalArgumentException("[neoDeploy] Invalid deployMode = '${deployMode}'. Valid 'deployMode' values are: 'MTA', 'WAR_PARAMS' and 'WAR_PROPERTIESFILE'")
        }

        def propertiesFile
        def warAction
        if (deployMode == 'WAR_PROPERTIESFILE' || deployMode == 'WAR_PARAMS') {
            warAction = utils.getMandatoryParameter(parameters, 'warAction', 'deploy')
            if (warAction != 'warAction' && warAction != 'deploy') {
                throw new IllegalArgumentException("[neoDeploy] Invalid warAction = '${warAction}'. Valid 'warAction' values are: 'deploy' and 'rolling-update'.")
            }
        }
        if (deployMode == 'WAR_PROPERTIESFILE') {
            propertiesFile = new File(utils.getMandatoryParameter(parameters, 'propertiesFile', null))
            if (!propertiesFile.isAbsolute()) {
                propertiesFile = new File(pwd(), propertiesFile.getPath())
            }
            if (!propertiesFile.exists()){
                error "Properties file cannot be found with parameter propertiesFile: '${propertiesFile}'."
            }
        }

        def applicationName
        def runtime
        def runtimeVersion
        def vmSize
        if (deployMode == 'WAR_PARAMS') {
            applicationName = utils.getMandatoryParameter(parameters, 'applicationName', null)
            runtime = utils.getMandatoryParameter(parameters, 'runtime', null)
            runtimeVersion = utils.getMandatoryParameter(parameters, 'runtimeVersion', null)
            vmSize = utils.getMandatoryParameter(parameters, 'vmSize', 'lite')
            if (vmSize != 'lite' && vmSize !='pro' && vmSize != 'prem' && vmSize != 'prem-plus') {
                throw new IllegalArgumentException("[neoDeploy] Invalid vmSize = '${vmSize}'. Valid 'vmSize' values are: 'lite', 'pro', 'prem' and 'prem-plus'.")
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

        if (deployMode.equals('MTA') || deployMode.equals('WAR_PARAMS')) {
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
                   --source "${archivePath.getAbsolutePath()}" \
                """

            if (deployMode == 'MTA') {
                sh """#!/bin/bash
                      "${neoExecutable}" deploy-mta \
                      ${commonDeployParams} \
                      --host '${deployHost}' \
                      --account '${deployAccount}' \
                      --synchronous
                   """
            }

            if (deployMode == 'WAR_PARAMS') {
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

            if (deployMode == 'WAR_PROPERTIESFILE') {
                sh """#!/bin/bash
                      "${neoExecutable}" ${warAction} \
                      ${commonDeployParams} \
                      ${propertiesFile.getAbsolutePath()}
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
