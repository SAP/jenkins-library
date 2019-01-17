import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationLoader
import com.sap.piper.k8s.ContainerMap
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = 'piperStageWrapper'

void call(Map parameters = [:], body) {

    final script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    def stageName = parameters.stageName?:env.STAGE_NAME

    // load default & individual configuration
    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixin(ConfigurationLoader.defaultStageConfiguration(this, stageName))
        .mixinGeneralConfig(script.commonPipelineEnvironment)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName)
        .mixin(parameters)
        .addIfEmpty('stageName', stageName)
        .dependingOn('stageName').mixin('ordinal')
        .use()

    stageLocking(config) {
        def containerMap = ContainerMap.instance.getMap().get(stageName) ?: [:]
        if (Boolean.valueOf(env.ON_K8S) && containerMap.size() > 0) {
            withEnv(["POD_NAME=${stageName}"]) {
                dockerExecuteOnKubernetes(script: script, containerMap: containerMap) {
                    executeStage(script, body, stageName, config, utils)
                }
            }
        } else {
            node(config.nodeLabel) {
                executeStage(script, body, stageName, config, utils)
            }
        }
    }
}

private void stageLocking(Map config, Closure body) {
    if (config.stageLocking) {
        lock(resource: "${env.JOB_NAME}/${config.ordinal}", inversePrecedence: true) {
            milestone config.ordinal
            body()
        }
    } else {
        body()
    }
}

private void executeStage(script, originalStage, stageName, config, utils) {

    boolean projectExtensions
    boolean globalExtensions
    def startTime = System.currentTimeMillis()

    try {
        //Add general stage stashes to config.stashContent
        config.stashContent += script.commonPipelineEnvironment.configuration.stageStashes?.get(stageName)?.unstash ?: []

        utils.unstashAll(config.stashContent)

        /* Defining the sources where to look for a project extension and a repository extension.
        * Files need to be named like the executed stage to be recognized.
        */
        def projectInterceptorFile = "${config.projectExtensionsDirectory}${stageName}.groovy"
        def globalInterceptorFile = "${config.globalExtensionsDirectory}${stageName}.groovy"
        projectExtensions = fileExists(projectInterceptorFile)
        globalExtensions = fileExists(globalInterceptorFile)
        // Pre-defining the real originalStage in body variable, might be overwritten later if extensions exist
        def body = originalStage

        // First, check if a global extension exists via a dedicated repository
        if (globalExtensions) {
            Script globalInterceptorScript = load(globalInterceptorFile)
            echo "[${STEP_NAME}] Found global interceptor '${globalInterceptorFile}' for ${stageName}."
            // If we call the global interceptor, we will pass on originalStage as parameter
            body = {
                globalInterceptorScript(script: script, originalStage: body, stageName: stageName, config: config)
            }
        }

        // Second, check if a project extension (within the same repository) exists
        if (projectExtensions) {
            Script projectInterceptorScript = load(projectInterceptorFile)
            echo "[${STEP_NAME}] Running project interceptor '${projectInterceptorFile}' for ${stageName}."
            // If we call the project interceptor, we will pass on body as parameter which contains either originalStage or the repository interceptor
            projectInterceptorScript(script: script, originalStage: body, stageName: stageName, config: config)
        } else {
            //TODO: assign projectInterceptorScript to body as done for globalInterceptorScript, currently test framework does not seem to support this case. Further investigations needed.
            body()
        }

    } finally {
        echo "Current build result in stage $stageName is ${script.currentBuild.currentResult}."
        //Perform stashing of selected files in workspace
        utils.stashList(script, script.commonPipelineEnvironment.configuration.stageStashes?.get(stageName)?.stashes ?: [])
        deleteDir()

        def duration = System.currentTimeMillis() - startTime
        utils.pushToSWA([eventType: 'library-os-stage', stageName: stageName, stepParam1: "${script.currentBuild.currentResult}", stepParam2: "${startTime}", stepParam3: "${duration}", stepParam4: "${projectExtensions}", stepParam5: "${globalExtensions}"], config)
    }
}
