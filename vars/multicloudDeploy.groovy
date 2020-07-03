import com.sap.piper.GenerateDocumentation
import com.sap.piper.CloudPlatform
import com.sap.piper.DeploymentType
import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import com.sap.piper.JenkinsUtils
import com.sap.piper.k8s.ContainerMap

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /** Defines the targets to deploy on Cloud Foundry.*/
    'cfTargets',
    /** Defines the targets to deploy on neo.*/
    'neoTargets',
    /** Executes the deployments in parallel.*/
    'parallelExecution'
]

@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Defines Cloud Foundry service instances to create as part of the deployment.
     * This is a _list_ of _objects_ with the following properties each:
     * - apiEndpoint
     * - credentialsId
     * - serviceManifest
     * - manifestVariablesFiles
     * - org
     * - space
     */
    'cfCreateServices',
    /** Defines the deployment type.*/
    'enableZeroDowntimeDeployment',
])

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /** The source file to deploy to SAP Cloud Platform.*/
    'source'
])

@Field Map CONFIG_KEY_COMPATIBILITY = [parallelExecution: 'features/parallelTestExecution']

/**
 * Deploys an application to multiple platforms (Cloud Foundry, SAP Cloud Platform) or to multiple instances of multiple platforms or the same platform.
 */
@GenerateDocumentation
void call(parameters = [:]) {

    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this
        def utils = parameters.utils ?: new Utils()
        def jenkinsUtils = parameters.jenkinsUtils ?: new JenkinsUtils()
        def stageName = parameters.stage ?: env.STAGE_NAME

        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixin(parameters, PARAMETER_KEYS)

        Map config = configHelper.use()

        configHelper
            .withMandatoryProperty('source', null, { config.neoTargets })

        utils.pushToSWA([
            step         : STEP_NAME,
            stepParamKey1: 'enableZeroDowntimeDeployment',
            stepParam1   : config.enableZeroDowntimeDeployment
        ], config)

        def index = 1
        def deployments = [:]

        if (config.cfCreateServices) {
            def createServices = [:]
            for (int i = 0; i < config.cfCreateServices.size(); i++) {
                Map createServicesConfig = config.cfCreateServices[i]
                createServices["Service Creation ${i + 1}"] = {
                    cloudFoundryCreateService(
                        script: script,
                        cloudFoundry: [
                            apiEndpoint           : createServicesConfig.apiEndpoint,
                            credentialsId         : createServicesConfig.credentialsId,
                            serviceManifest       : createServicesConfig.serviceManifest,
                            manifestVariablesFiles: createServicesConfig.manifestVariablesFiles,
                            org                   : createServicesConfig.org,
                            space                 : createServicesConfig.space
                        ]
                    )
                }
            }
            runClosures(config, createServices, "cloudFoundryCreateService")
        }

        if (config.cfTargets) {

            def deploymentType = DeploymentType.selectFor(CloudPlatform.CLOUD_FOUNDRY, config.enableZeroDowntimeDeployment).toString()

            for (int i = 0; i < config.cfTargets.size(); i++) {

                def target = config.cfTargets[i]

                // An isolated workspace is only required when using blue-green deployment with multiple cfTargets,
                // since the cloudFoundryDeploy step might edit the manifest.yml file in that case.
                Boolean runInIsolatedWorkspace = config.cfTargets.size() > 1 && (deploymentType == "blue-green" || (target.mtaExtensionCredentials && config.parallelExecution))

                if (target.mtaExtensionCredentials) runInIsolatedWorkspace = true

                Closure deployment = {
                    Utils deploymentUtils = new Utils()
                    if (runInIsolatedWorkspace) {
                        deploymentUtils.unstashStageFiles(script, stageName)
                    }

                    cloudFoundryDeploy(
                        script: script,
                        juStabUtils: utils,
                        jenkinsUtilsStub: jenkinsUtils,
                        deployType: deploymentType,
                        cloudFoundry: target,
                        mtaPath: script.commonPipelineEnvironment.mtarFilePath,
                    )
                    if (runInIsolatedWorkspace) {
                        deploymentUtils.stashStageFiles(script, stageName)
                    }
                }
                if (runInIsolatedWorkspace){
                    deployments["Deployment ${index}"] = {
                        if (env.POD_NAME) {
                            dockerExecuteOnKubernetes(script: script, containerMap: ContainerMap.instance.getMap().get(stageName) ?: [:]) {
                                deployment.call()
                            }
                        } else {
                            node(env.NODE_NAME) {
                                deployment.call()
                            }
                        }
                    }
                } else {
                    deployments.put("Deployment ${index}", deployment)
                }
                index++
            }
        }

        if (config.neoTargets) {

            def deploymentType = DeploymentType.selectFor(CloudPlatform.NEO, config.enableZeroDowntimeDeployment).toString()

            for (int i = 0; i < config.neoTargets.size(); i++) {

                def target = config.neoTargets[i]

                Closure deployment = {

                    neoDeploy(
                        script: script,
                        warAction: deploymentType,
                        source: config.source,
                        neo: target
                    )

                }
                deployments.put("Deployment ${index}", deployment)
                index++
            }
        }

        if (!config.cfTargets && !config.neoTargets) {
            error "Deployment skipped because no targets defined!"
        }

        runClosures(config, deployments, "deployments")

    }
}

def runClosures(Map config, Map toRun, String label = "closures") {
    echo "Executing $label"
    if (config.parallelExecution) {
        echo "Executing $label in parallel"
        parallel toRun
    } else {
        echo "Executing $label in sequence"
        def closuresToRun = toRun.values().asList()
        for (int i = 0; i < closuresToRun.size(); i++) {
            (closuresToRun[i] as Closure)()
        }
    }
}
