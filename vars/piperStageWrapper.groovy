import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationLoader
import com.sap.piper.DebugReport
import com.sap.piper.k8s.ContainerMap
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = 'piperStageWrapper'

void call(Map parameters = [:], body) {

    final script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    def stageName = parameters.stageName ?: env.STAGE_NAME

    // load default & individual configuration
    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixin(ConfigurationLoader.defaultStageConfiguration(script, stageName))
        .mixinGeneralConfig(script.commonPipelineEnvironment)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName)
        .mixin(parameters)
        .addIfEmpty('stageName', stageName)
        .dependingOn('stageName').mixin('ordinal')
        .use()

    stageLocking(config) {
        // failOnError needs to be set to not harm resilience feature
        // if not set to true: failures in mandatory steps will be caught here and neglected
        handlePipelineStepErrors(stepName: stageName, stepParameters: parameters, failOnError: true) {
            def containerMap = ContainerMap.instance.getMap().get(stageName) ?: [:]
            if (Boolean.valueOf(env.ON_K8S) && containerMap.size() > 0) {
                DebugReport.instance.environment.put("environment", "Kubernetes")
                withEnv(["POD_NAME=${stageName}"]) {
                    dockerExecuteOnKubernetes(script: script, containerMap: containerMap, stageName: stageName) {
                        executeStage(script, body, stageName, config, utils, parameters.telemetryDisabled)
                    }
                }
            } else {
                node(config.nodeLabel) {
                    executeStage(script, body, stageName, config, utils, parameters.telemetryDisabled)
                }
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

private void executeStage(script, originalStage, stageName, config, utils, telemetryDisabled = false) {
    boolean projectExtensions
    boolean globalExtensions
    def startTime = System.currentTimeMillis()

    try {
        // Add general stage stashes to config.stashContent
        config.stashContent = utils.unstashStageFiles(script, stageName, config.stashContent)

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
        if (globalExtensions && allowExtensions()) {
            echo "[${STEP_NAME}] Found global interceptor '${globalInterceptorFile}' for ${stageName}."
            // If we call the global interceptor, we will pass on originalStage as parameter
            DebugReport.instance.globalExtensions.put(stageName, "Overwrites")
            Closure modifiedOriginalStage = {
                DebugReport.instance.globalExtensions.put(stageName, "Extends")
                originalStage()
            }

            body = {
                callInterceptor(script, globalInterceptorFile, modifiedOriginalStage, stageName, config)
            }
        }

        // Second, check if a project extension (within the same repository) exists
        if (projectExtensions && allowExtensions()) {
            echo "[${STEP_NAME}] Running project interceptor '${projectInterceptorFile}' for ${stageName}."
            // If we call the project interceptor, we will pass on body as parameter which contains either originalStage or the repository interceptor
            if (projectExtensions && globalExtensions) {
                DebugReport.instance.globalExtensions.put(stageName, "Unknown (Overwritten by local extension)")
            }
            DebugReport.instance.localExtensions.put(stageName, "Overwrites")
            Closure modifiedOriginalBody = {
                DebugReport.instance.localExtensions.put(stageName, "Extends")
                if (projectExtensions && globalExtensions) {
                    DebugReport.instance.globalExtensions.put(stageName, "Overwrites")
                }
                body.call()
            }

            callInterceptor(script, projectInterceptorFile, modifiedOriginalBody, stageName, config)

        } else {
            // NOTE: It may appear more elegant to re-assign 'body' more than once and then call 'body()' after the
            // if-block. This could lead to infinite loops however, as any change to the local variable 'body' will
            // become visible in all of the closures at the time they run. I.e. 'body' inside any of the closures will
            // reflect the last assignment and not its value at the time of constructing the closure!
            body()
        }

    } finally {
        //Perform stashing of selected files in workspace
        utils.stashStageFiles(script, stageName)

        // In general telemetry reporting is disabled by the config settings. This flag is used to disable the reporting when the config is not yet read (e.g. init stage).
        if(!telemetryDisabled){
            def duration = System.currentTimeMillis() - startTime
            utils.pushToSWA([
                eventType: 'library-os-stage',
                stageName: stageName,
                stepParamKey1: 'buildResult',
                stepParam1: "${script.currentBuild.currentResult}",
                buildResult: "${script.currentBuild.currentResult}",
                stepParamKey2: 'stageStartTime',
                stepParam2: "${startTime}",
                stageStartTime: "${startTime}",
                stepParamKey3: 'stageDuration',
                stepParam3: "${duration}",
                stageDuration: "${duration}",
                stepParamKey4: 'projectExtension',
                stepParam4: "${projectExtensions}",
                projectExtension: "${projectExtensions}",
                stepParamKey5: 'globalExtension',
                stepParam5: "${globalExtensions}",
                globalExtension: "${globalExtensions}"
            ], config)
        }
    }
}

private void callInterceptor(Script script, String extensionFileName, Closure originalStage, String stageName, Map configuration) {
    Script interceptor = load(extensionFileName)
    if (isOldInterceptorInterfaceUsed(interceptor)) {
        echo("[Warning] The interface to implement extensions has changed. " +
            "The extension $extensionFileName has to implement a method named 'call' with exactly one parameter of type Map. " +
            "This map will have the properties script, originalStage, stageName, config. " +
            "For example: def call(Map parameters) { ... }")
        interceptor.call(originalStage, stageName, configuration, configuration)
    } else {
        validateInterceptor(interceptor, extensionFileName)
        interceptor.call([
            script       : script,
            originalStage: originalStage,
            stageName    : stageName,
            config       : configuration
        ])
    }
}

@NonCPS
private boolean isInterceptorValid(Script interceptor) {
    MetaMethod method = interceptor.metaClass.pickMethod("call", [Map.class] as Class[])
    return method != null
}

private void validateInterceptor(Script interceptor, String extensionFileName) {
    if (!isInterceptorValid(interceptor)) {
        error("The extension $extensionFileName has to implement a method named 'call' with exactly one parameter of type Map. " +
            "This map will have the properties script, originalStage, stageName, config. " +
            "For example: def call(Map parameters) { ... }")
    }
}

@NonCPS
private boolean isOldInterceptorInterfaceUsed(Script interceptor) {
    MetaMethod method = interceptor.metaClass.pickMethod("call", [Closure.class, String.class, Map.class, Map.class] as Class[])
    return method != null
}

private boolean allowExtensions(){
    return env.PIPER_DISABLE_EXTENSIONS == null || Boolean.valueOf(env.PIPER_DISABLE_EXTENSIONS) == false
}
