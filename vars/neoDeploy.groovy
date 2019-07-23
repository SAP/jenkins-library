import com.sap.piper.GenerateDocumentation
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
    'neo',
    /**
     * The SAP Cloud Platform account to deploy to.
     * @parentConfigKey neo
     * @mandatory for deployMode=warParams
     */
    'account',
    /**
     * Name of the application you want to manage, configure, or deploy.
     * @parentConfigKey neo
     * @mandatory for deployMode=warParams
     */
    'application',
    /**
     * The Jenkins credentials containing user and password used for SAP CP deployment.
     * @parentConfigKey neo
     */
    'credentialsId',
    /**
     * Map of environment variables in the form of KEY: VALUE.
     * @parentConfigKey neo
     */
    'environment',
    /**
     * The SAP Cloud Platform host to deploy to.
     * @parentConfigKey neo
     * @mandatory for deployMode=warParams
     */
    'host',
        /**
     * The path to the .properties file in which all necessary deployment properties for the application are defined.
     * @parentConfigKey neo
     * @mandatory for deployMode=warPropertiesFile
     */
    'propertiesFile',
    /**
     * Name of SAP Cloud Platform application runtime.
     * @parentConfigKey neo
     * @mandatory for deployMode=warParams
     */
    'runtime',
    /**
     * Version of SAP Cloud Platform application runtime.
     * @parentConfigKey neo
     * @mandatory for deployMode=warParams
     */
    'runtimeVersion',
        /**
     * Compute unit (VM) size. Acceptable values: lite, pro, prem, prem-plus.
     * @parentConfigKey neo
     */
    'size',
        /**
     * String of VM arguments passed to the JVM.
     * @parentConfigKey neo
     */
    'vmArguments'
]

@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * @see dockerExecute
     */
    'dockerEnvVars',
    /**
     * @see dockerExecute
     */
    'dockerImage',
    /**
     * @see dockerExecute
     */
    'dockerOptions',
    /**
      * Extension files. Provided to the neo command via parameter `--extensions` (`-e`). Only valid for deploy mode `mta`.
      */
    'extensions',
    /**
     * The path to the archive for deployment to SAP CP. If not provided `mtarFilePath` from commom pipeline environment is used instead.
     */
    'source'
])

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /**
     * The deployment mode which should be used. Available options are:
     * *`'mta'` - default,
     * *`'warParams'` - deploying WAR file and passing all the deployment parameters via the function call,
     * *`'warPropertiesFile'` - deploying WAR file and putting all the deployment parameters in a .properties file.
     * @possibleValues 'mta', 'warParams', 'warPropertiesFile'
     */
    'deployMode',
    /**
     * Action mode when using WAR file mode. Available options are `deploy` (default) and `rolling-update` which performs update of an application without downtime in one go.
     * @possibleValues 'deploy', 'rolling-update'
     */
    'warAction'
])

/**
 * Deploys an Application to SAP Cloud Platform (SAP CP) using the SAP Cloud Platform Console Client (Neo Java Web SDK).
 */
@GenerateDocumentation
void call(parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this

        def utils = parameters.utils ?: new Utils()

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

        Set extensionFileNames

        if(configuration.extensions == null) {
            extensionFileNames = []
        } else {
            extensionFileNames = configuration.extensions in Collection ? configuration.extensions : [configuration.extensions]
        }

        if( ! extensionFileNames.findAll { it == null || it.isEmpty() }.isEmpty() )
            error "At least one extension file name was null or empty: ${extensionFileNames}."

        if(deployMode != DeployMode.MTA && ! extensionFileNames.isEmpty())
            error "Extensions (${extensionFileNames} found for deploy mode ${deployMode}. Extensions are only supported for deploy mode '${DeployMode.MTA}')"

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

                for(CharSequence extensionFile in extensionFileNames) {
                    StepAssertions.assertFileExists(this, extensionFile)
                }

                NeoCommandHelper neoCommandHelper = new NeoCommandHelper(
                    this,
                    deployMode,
                    configuration.neo,
                    extensionFileNames,
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

    String logFolder = 'logs/neo'

    try {
        sh "mkdir -p ${logFolder}"
        withEnv(["neo_logging_location=${pwd()}/${logFolder}"]) {
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
                    def returnCodeRollingUpdate = sh returnStatus: true, script: neoCommandHelper.rollingUpdateCommand()
                    if(returnCodeRollingUpdate != 0){
                        error "[ERROR][${STEP_NAME}] The execution of the deploy command failed, see the log for details."
                    }
                } else {
                    def returnCodeDeploy = sh returnStatus: true, script: neoCommandHelper.deployCommand()
                    if(returnCodeDeploy != 0){
                        error "[ERROR][${STEP_NAME}] The execution of the deploy command failed, see the log for details."
                    }
                    sh neoCommandHelper.restartCommand()
                }


            } else if (deployMode == DeployMode.MTA) {
                def returnCodeMTA = sh returnStatus: true, script: neoCommandHelper.deployMta()
                if(returnCodeMTA != 0){
                    error "[ERROR][${STEP_NAME}] The execution of the deploy command failed, see the log for details."
                }
            }
        }
    }
    catch (Exception ex) {

        echo "Error while deploying to SAP Cloud Platform. Here are the neo.sh logs:"
        try {
            sh "cat ${logFolder}/*"
        } catch(Exception e) {
            echo "Unable to provide the logs."
            ex.addSuppressed(e)
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
