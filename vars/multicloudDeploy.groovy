import com.sap.piper.GenerateDocumentation
import com.sap.piper.CloudPlatform
import com.sap.piper.DeploymentType
import com.sap.piper.k8s.ContainerMap
import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import com.sap.piper.JenkinsUtils

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /** Defines the targets to deploy on cloudFoundry.*/
    'cfTargets',
    /** Defines the targets to deploy on neo.*/
    'neoTargets'
]

@Field Set STEP_CONFIG_KEYS = []

@Field Set PARAMETER_KEYS = GENERAL_CONFIG_KEYS.plus([
    /** The stage name. If the stage name is not provided, it will be taken from the environment variable 'STAGE_NAME'.*/
    'stage',
    /** Defines the deployment type.*/
    'enableZeroDowntimeDeployment',
    /** The source file to deploy to the SAP Cloud Platform.*/
    'source'
])

/**
 * Deploys an application to multiple platforms (cloudFoundry, SAP Cloud Platform) or to multiple instances of multiple platforms or the same platform.
 */
@GenerateDocumentation
void call(parameters = [:]) {

    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        def stageName = parameters.stage ?: env.STAGE_NAME
        def enableZeroDowntimeDeployment = parameters.enableZeroDowntimeDeployment ?: false

        def script = checkScript(this, parameters) ?: this
        def utils = parameters.utils ?: new Utils()
        def jenkinsUtils = parameters.jenkinsUtils ?: new JenkinsUtils()

        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this, script)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)

        Map config = configHelper.use()

        configHelper
            .withMandatoryProperty('source', null, { config.neoTargets })

        utils.pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'stage',
            stepParam1: stageName,
            stepParamKey2: 'enableZeroDowntimeDeployment',
            stepParam2: enableZeroDowntimeDeployment
        ], config)

        def index = 1
        def deployments = [:]
        def deploymentType
        def deployTool

        if (config.cfTargets) {

            deploymentType = DeploymentType.selectFor(CloudPlatform.CLOUD_FOUNDRY, enableZeroDowntimeDeployment).toString()
            deployTool = script.commonPipelineEnvironment.configuration.isMta ? 'mtaDeployPlugin' : 'cf_native'

            for (int i = 0; i < config.cfTargets.size(); i++) {

                def target = config.cfTargets[i]

                Closure deployment = {

                    cloudFoundryDeploy(
                        script: script,
                        juStabUtils: utils,
                        jenkinsUtilsStub: jenkinsUtils,
                        deployType: deploymentType,
                        cloudFoundry: target,
                        mtaPath: script.commonPipelineEnvironment.mtarFilePath,
                        deployTool: deployTool
                    )

                }
                setDeployment(deployments, deployment, index, script, stageName)
                index++
            }
            utils.runClosures(deployments)
        }

        if (config.neoTargets) {

            deploymentType = DeploymentType.selectFor(CloudPlatform.NEO, enableZeroDowntimeDeployment)

            for (int i = 0; i < config.neoTargets.size(); i++) {

                def target = config.neoTargets[i]

                Closure deployment = {

                    neoDeploy (
                        script: script,
                        warAction: deploymentType.toString(),
                        source: config.source,
                        neo: target
                    )

                }
                setDeployment(deployments, deployment, index, script, stageName)
                index++
            }
            utils.runClosures(deployments)
        }

        if (!config.cfTargets && !config.neoTargets) {
            error "Deployment skipped because no targets defined!"
        }
    }
}

void setDeployment(deployments, deployment, index, script, stageName) {
    deployments["Deployment ${index > 1 ? index : ''}"] = {
        if (env.POD_NAME) {
            dockerExecuteOnKubernetes(script: script, containerMap: ContainerMap.instance.getMap().get(stageName) ?: [:]) {
                deployment.run()
            }
        } else {
            node(env.NODE_NAME) {
                deployment.run()
            }
        }
    }
}
