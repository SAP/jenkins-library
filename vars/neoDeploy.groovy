import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import com.sap.piper.StepAssertions
import com.sap.piper.tools.neo.DeployMode
import com.sap.piper.tools.neo.NeoCommandHelper
import com.sap.piper.tools.neo.WarAction
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = [
    'neo'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'dockerEnvVars',
    'dockerImage',
    'dockerOptions',
    'neoHome',
    'source'
])

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    'deployMode',
    'warAction'
])

void call(parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this

        def utils = parameters.utils ?: new Utils()

        prepareDefaultValues script: script

        // load default & individual configuration
        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName ?: env.STAGE_NAME, STEP_CONFIG_KEYS)
            .addIfEmpty('source', script.commonPipelineEnvironment.getMtarFilePath())
            .mixin(parameters, PARAMETER_KEYS)
            .collectValidationFailures()
            .withPropertyInValues('deployMode', DeployMode.stringValues())

        Map configuration = configHelper.use()

        DeployMode deployMode = DeployMode.fromString(configuration.deployMode)

        def isWarParamsDeployMode = { deployMode == DeployMode.WAR_PARAMS },
            isNotWarPropertiesDeployMode = {deployMode != DeployMode.WAR_PROPERTIES_FILE}

        configHelper
            .withMandatoryProperty('source')
            .withMandatoryProperty('neo/credentialsId')
            .withMandatoryProperty('neo/application', null, isWarParamsDeployMode)
            .withMandatoryProperty('neo/runtime', null, isWarParamsDeployMode)
            .withMandatoryProperty('neo/runtimeVersion', null, isWarParamsDeployMode)
            .withMandatoryProperty('neo/host', null, isNotWarPropertiesDeployMode)
            .withMandatoryProperty('neo/account', null, isNotWarPropertiesDeployMode)
            //
            // call 'use()' a second time in order to get the collected validation failures
            // since the map did not change, it is not required to replace the previous configuration map.
            .use()

        utils.pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'deployMode',
            stepParam1: configuration.deployMode == 'mta'?'mta':'war', // ['mta', 'warParams', 'warPropertiesFile']
            stepParamKey2: 'warAction',
            stepParam2: configuration.warAction == 'rolling-update'?'blue-green':'standard', // ['deploy', 'deploy-mta', 'rolling-update']
            stepParamKey3: 'scriptMissing',
            stepParam3: parameters?.script == null,
        ], configuration)


        withCredentials([usernamePassword(
            credentialsId: configuration.neo.credentialsId,
            passwordVariable: 'NEO_PASSWORD',
            usernameVariable: 'NEO_USERNAME')]) {

            assertPasswordRules(NEO_PASSWORD)

            dockerExecute(
                script: script,
                dockerImage: configuration.dockerImage,
                dockerEnvVars: configuration.dockerEnvVars,
                dockerOptions: configuration.dockerOptions
            ) {

                StepAssertions.assertFileExists(this, configuration.source)

                NeoCommandHelper neoCommandHelper = new NeoCommandHelper(
                    this,
                    deployMode,
                    configuration.neo,
                    NEO_USERNAME,
                    NEO_PASSWORD,
                    configuration.source
                )

                lock("$STEP_NAME :${neoCommandHelper.resourceLock()}") {
                    deploy(script, utils, configuration, neoCommandHelper, configuration.dockerImage, deployMode)
                }
            }
        }
    }
}

private deploy(script, utils, Map configuration, NeoCommandHelper neoCommandHelper, dockerImage, DeployMode deployMode) {

    try {
        sh "mkdir -p logs/neo"
        withEnv(["neo_logging_location=${pwd()}/logs/neo"]) {
            if (deployMode.isWarDeployment()) {
                ConfigurationHelper.newInstance(this, configuration).withPropertyInValues('warAction', WarAction.stringValues())
                WarAction warAction = WarAction.fromString(configuration.warAction)

                if (warAction == WarAction.ROLLING_UPDATE) {
                    if (!isAppRunning(neoCommandHelper)) {
                        warAction = WarAction.DEPLOY
                        echo "Rolling update not possible because application is not running. Falling back to standard deployment."
                    }
                }

                echo "Link to the application dashboard: ${neoCommandHelper.cloudCockpitLink()}"

                if (warAction == WarAction.ROLLING_UPDATE) {
                    sh neoCommandHelper.rollingUpdateCommand()
                } else {
                    sh neoCommandHelper.deployCommand()
                    sh neoCommandHelper.restartCommand()
                }


            } else if (deployMode == DeployMode.MTA) {
                sh neoCommandHelper.deployMta()
            }
        }
    }
    catch (Exception ex) {
        if (dockerImage) {
            echo "Error while deploying to SAP Cloud Platform. Here are the neo.sh logs:"
            sh "cat logs/neo/*"
        }
        throw ex
    }
}

private boolean isAppRunning(NeoCommandHelper commandHelper) {
    def status = sh script: "${commandHelper.statusCommand()} || true", returnStdout: true
    return status.contains('Status: STARTED')
}

private assertPasswordRules(String password) {
    if (password.startsWith("@")) {
        error("Your password for the deployment to SAP Cloud Platform contains characters which are not " +
            "supported by the neo tools. " +
            "For example it is not allowed that the password starts with @. " +
            "Please consult the documentation for the neo command line tool for more information: " +
            "https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/8900b22376f84c609ee9baf5bf67130a.html")
    }
}
