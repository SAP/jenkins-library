import com.sap.piper.tools.neo.NeoCommandHelper

import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils

import com.sap.piper.tools.ToolDescriptor

import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
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
    'deployMode',
    'propertiesFile',
    'runtime',
    'runtimeVersion',
    'vmSize',
    'vmArguments',
    'environment',
    'warAction'
])

void call(parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this

        def utils = parameters.utils ?: new Utils()

        prepareDefaultValues script: script

        Map stepCompatibilityConfiguration = handleCompatibility(script, parameters)

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
            stepParam3: parameters?.script == null,
            stepParam4: ! stepCompatibilityConfiguration.isEmpty(),
        ], configuration)

        ToolDescriptor neo = new ToolDescriptor('SAP Cloud Platform Console Client', 'NEO_HOME', 'neoHome', '/tools/', 'neo.sh', null, 'version')
        ToolDescriptor java = new ToolDescriptor('Java', 'JAVA_HOME', '', '/bin/', 'java', '1.8.0', '-version 2>&1')

        if(configuration.neoCredentialsId) {
            withCredentials([usernamePassword(
                credentialsId: credentialsId,
                passwordVariable: 'NEO_PASSWORD',
                usernameVariable: 'NEO_USERNAME')]) {

                assertPasswordRules(NEO_PASSWORD)
                NeoCommandHelper neoCommandHelper = new NeoCommandHelper(script, configuration, neo, NEO_USERNAME, NEO_PASSWORD)

                dockerExecute(
                    script: script,
                    dockerImage: configuration.get('dockerImage'),
                    dockerEnvVars: configuration.get('dockerEnvVars'),
                    dockerOptions: configuration.get('dockerOptions')
                ) {

                    neo.verify(this, configuration)
                    java.verify(this, configuration)

                    lock("$STEP_NAME :${neoCommandHelper.resourceLock()}") {
                        deploy(script, utils, configuration, neoCommandHelper, neo)
                    }
                }
            }
        }
        else {
            error("[neoDeploy] No credentials defined for the deployment. Please specify the value for neoCredentialsId.")
        }
    }
}

private deploy(script, utils, Map configuration, NeoCommandHelper neoCommandHelper, ToolDescriptor neoToolDescriptor){
    def deployModes = ['mta', 'warParams', 'warPropertiesFile']
    def deployMode = utils.getParameterInValueRange(script, configuration, 'deployMode', deployModes)

    try {
        if (deployMode in ['warPropertiesFile', 'warParams']) {
            def warActions = ['deploy', 'rolling-update']
            def warAction = utils.getParameterInValueRange(script, configuration, 'warAction', warActions)

            if (warAction == 'rolling-update') {
                if (!isAppRunning(neoCommandHelper)) {
                    warAction = 'deploy'
                    echo "Rolling update not possible because application is not running. Falling back to standard deployment."
                }
            }

            echo "Link to the application dashboard: ${neoCommandHelper.cloudCockpitLink()}"

            if (warAction == 'rolling-update') {
                sh neoCommandHelper.rollingUpdateCommand()
            } else {
                sh neoCommandHelper.deployCommand()
                sh neoCommandHelper.restartCommand()
            }


        } else if (deployMode == 'mta') {
            warAction = 'deploy-mta'

            sh neoCommandHelper.deployMta()
        }
    }
    catch (Exception ex) {
        echo "Error while deploying to SAP Cloud Platform. Here are the neo.sh logs:"
        sh "cat ${neoToolDescriptor.getToolLocation()}/tools/log/*"
        throw ex
    }
}

private boolean isAppRunning(NeoCommandHelper commandHelper) {
    def status = sh script: "${commandHelper.statusCommand()} || true", returnStdout: true
    return status.contains('Status: STARTED')
}

private handleCompatibility(script, parameters){
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

    if(! stepCompatibilityConfiguration.isEmpty()) {
        echo "[WARNING][$STEP_NAME] You are using a deprecated configuration framework. This will be removed in " +
            'futureVersions.\nAdd snippet below to \'./pipeline/config.yml\' and remove ' +
            'file \'.pipeline/configuration.properties\'.\n' +
            """|steps:
                    |    neoDeploy:
                    |        host: ${stepCompatibilityConfiguration.get('host', '<Add host here>')}
                    |        account: ${stepCompatibilityConfiguration.get('account', '<Add account here>')}
                """.stripMargin()

        if(Boolean.getBoolean('com.sap.piper.featureFlag.buildUnstableWhenOldConfigFrameworkIsUsedByNeoDeploy')) {
            script.currentBuild.setResult('UNSTABLE')
            echo "[WARNING][$STEP_NAME] Build has been set to unstable since old config framework is used."
        }
    }

    return stepCompatibilityConfiguration
}

private assertPasswordRules(String password){
    if(password.startsWith("@")){
        error("Your password for the deployment to SAP Cloud Platform contains characters which are not " +
            "supported by the neo tools. " +
            "For example it is not allowed that the password starts with @. " +
            "Please consult the documentation for the neo command line tool for more information: " +
            "https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/8900b22376f84c609ee9baf5bf67130a.html")
    }
}
