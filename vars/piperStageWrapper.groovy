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
        .addIfEmpty('lockingResourceGroup', script.commonPipelineEnvironment.projectName)
        .dependingOn('stageName').mixin('ordinal')
        .use()

    stageLocking(config) {
        def containerMap = ContainerMap.instance.getMap().get(stageName) ?: [:]
        List environment = []
        if (stageName && stageName != env.STAGE_NAME) {
            // Avoid two sources of truth with regards to stageName.
            // env.STAGE_NAME is filled from stage('Display name') {, it only serves the purpose of
            // easily getting to the stage name from steps.
            environment.add("STAGE_NAME=${stageName}")
        }
        if (config.sidecarImage) {
            echo "sidecarImage configured for stage '${stageName}': '${config.sidecarImage}'"
            environment.add("SIDECAR_IMAGE=${config.sidecarImage}")
        }
        if (Boolean.valueOf(env.ON_K8S) && (containerMap.size() > 0 || config.runStageInPod)) {
            environment.add("POD_NAME=${stageName}")
            withEnv(environment) {
                dockerExecuteOnKubernetes(script: script, containerMap: containerMap, stageName: stageName) {
                    executeStage(script, body, stageName, config, utils, parameters.telemetryDisabled)
                }
            }
        } else {
            withEnvWrapper(environment) {
                node(config.nodeLabel) {
                    executeStage(script, body, stageName, config, utils, parameters.telemetryDisabled)
                }
            }
        }
    }
}

private void withEnvWrapper(List environment, Closure body) {
    if (environment) {
        withEnv(environment) {
            body()
        }
    } else {
        body()
    }
}

private void stageLocking(Map config, Closure body) {
    if (config.stageLocking) {
        String resource = config.lockingResourceGroup?:env.JOB_NAME
        if(config.lockingResource){
            resource += "/${config.lockingResource}"
        }
        else if(config.ordinal){
            resource += "/${config.ordinal}"
        }
        lock(resource: resource, inversePrecedence: true) {
            if(config.ordinal) {
                milestone config.ordinal
            }
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
        /* due to renaming stage 'Central Build' to 'Build' need to define extension file name 'Central Build.groovy'
        as stageName used to generate it, once all the users will 'Build' as a stageName
        and extension filename, below renaming snippet should be removed
        */
        if (stageName == 'Build'){
            if (!fileExists(projectInterceptorFile) || !fileExists(globalInterceptorFile)){
                def centralBuildExtensionFileName = "Central Build.groovy"
                projectInterceptorFile = "${config.projectExtensionsDirectory}${centralBuildExtensionFileName}"
                globalInterceptorFile = "${config.globalExtensionsDirectory}${centralBuildExtensionFileName}"
            }
        }

        projectExtensions = fileExists(projectInterceptorFile)
        globalExtensions = fileExists(globalInterceptorFile)
        // Pre-defining the real originalStage in body variable, might be overwritten later if extensions exist
        def body = originalStage

        // First, check if a global extension exists via a dedicated repository
        if (globalExtensions && allowExtensions(script)) {
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
        if (projectExtensions && allowExtensions(script)) {
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
    }
}

private void callInterceptor(Script script, String extensionFileName, Closure originalStage, String stageName, Map configuration) {
    try {
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
    } catch (Throwable error) {
        if (!DebugReport.instance.failedBuild.step) {
            DebugReport.instance.storeStepFailure("${stageName}(extended)", error, true)
        }
        throw error
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

private static boolean allowExtensions(Script script) {
    return script.env.PIPER_DISABLE_EXTENSIONS == null || Boolean.valueOf(script.env.PIPER_DISABLE_EXTENSIONS) == false
}
