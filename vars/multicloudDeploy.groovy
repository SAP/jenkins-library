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
    /** Defines the deployment type.*/
    'enableZeroDowntimeDeployment',
    /** Executes the deployments in parallel.*/
    'parallelExecution',
    /** The source file to deploy to SAP Cloud Platform.*/
    'source'
])

/**
 * Deploys an application to multiple platforms (Cloud Foundry, SAP Cloud Platform) or to multiple instances of multiple platforms or the same platform.
 */
@GenerateDocumentation
void call(parameters = [:]) {

    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this
        def utils = parameters.utils ?: new Utils()
        def jenkinsUtils = parameters.jenkinsUtils ?: new JenkinsUtils()

        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)

        Map config = configHelper.use()

        configHelper
            .withMandatoryProperty('source', null, { config.neoTargets })

        utils.pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'enableZeroDowntimeDeployment',
            stepParam1: config.enableZeroDowntimeDeployment
        ], config)

        def index = 1
        def deployments = [:]

        if (config.cfTargets) {

            def deploymentType = DeploymentType.selectFor(CloudPlatform.CLOUD_FOUNDRY, config.enableZeroDowntimeDeployment).toString()
            def deployTool = script.commonPipelineEnvironment.configuration.isMta ? 'mtaDeployPlugin' : 'cf_native'

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
                setDeployment(deployments, deployment, index)
                index++
            }
        }

        if (config.neoTargets) {

            def deploymentType = DeploymentType.selectFor(CloudPlatform.NEO, config.enableZeroDowntimeDeployment).toString()

            for (int i = 0; i < config.neoTargets.size(); i++) {

                def target = config.neoTargets[i]

                Closure deployment = {

                    neoDeploy (
                        script: script,
                        warAction: deploymentType,
                        source: config.source,
                        neo: target
                    )

                }
                setDeployment(deployments, deployment, index)
                index++
            }
        }

        if (!config.cfTargets && !config.neoTargets) {
            error "Deployment skipped because no targets defined!"
        }

        runDeployments(utils, config.parallelExecution, deployments)
    }
}

void setDeployment(deployments, deployment, index) {
    echo "Setting up deployments"
    deployments["Deployment ${index}"] = {
        deployment.run()
    }
}

void runDeployments(utils, parallelExecution, deployments) {
    echo "Executing deployments"
    if (parallelExecution) {
        echo "Executing deployments in parallel"
        parallel deployments
    } else {
        echo "Executing deployments in sequence"
        utils.runClosures(deployments)
    }
}
