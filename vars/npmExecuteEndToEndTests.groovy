import com.sap.piper.ConfigurationHelper
import com.sap.piper.DownloadCacheUtils
import com.sap.piper.GenerateDocumentation
import com.sap.piper.k8s.ContainerMap
import groovy.transform.Field
import com.sap.piper.Utils

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /** Executes the deployments in parallel.*/
    'parallelExecution',
    /**
     * The branch used as productive branch, defaults to master.
     */
    'productiveBranch'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * The URLs under which the app is available after deployment.
     * Each element of appUrls must be a map containing a property url, an optional property credentialId, and an optional property parameters.
     * The optional property parameters can be used to pass additional parameters to the end-to-end test deployment reachable via the given application URL.
     * These parameters must be a list of strings, where each string corresponds to one element of the parameters.
     * For example, if the parameter `--tag scenario1` should be passed to the test, specify parameters: ["--tag", "scenario1"].
     * These parameters are appended to the npm command during execution.
     */
    'appUrls',
    /**
     * List of build descriptors and therefore modules to exclude from execution of the npm scripts.
     * The elements of the list can either be a path to the build descriptor or a pattern.
     */
    'buildDescriptorExcludeList',
    /**
     * Script to be executed from package.json. Defaults to `ci-e2e`.
     */
    'runScript',
    /**
     * Boolean to indicate whether the step should only be executed in the productive branch or not.
     * @possibleValues `true`, `false`
     */
    'onlyRunInProductiveBranch',
    /**
     * Docker image on which end to end tests should be executed
     */
    'dockerImage',
    /**
     * Base URL of the application to be tested
     */
    'baseUrl',
    /**
     * Credentials to access the application to be tested
     */
    'credentialsId',
    /**
     * Distinguish if these are wdi5 tests. If set to `true` `wdi5_username` and `wdi5_password` environment variables are used to enable [autologin](https://ui5-community.github.io/wdi5/#/authentication?id=credentials).
     * @possibleValues `true`, `false`
     */
    'wdi5'

])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

@Field Map CONFIG_KEY_COMPATIBILITY = [parallelExecution: 'features/parallelTestExecution']

/**
 * Executes end to end tests by running the npm script configured via the `runScript` property.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        String stageName = parameters.stageName ?: env.STAGE_NAME

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        // telemetry reporting
        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], config)

        def e2ETests = [:]
        def index = 1

        def npmParameters = [:]
        npmParameters.dockerOptions = ['--shm-size 512MB']

        if (config.appUrls && !(config.appUrls instanceof List)) {
            error "[${STEP_NAME}] The execution failed, since appUrls is not a list. Please provide appUrls as a list of maps. For example:\n" +
                    "appUrls: \n" + "  - url: 'https://my-url.com'\n" + "    credentialId: myCreds"

        }

        if (config.onlyRunInProductiveBranch && (config.productiveBranch != env.BRANCH_NAME)) {
            return
        }
        List credentials = []

        if (config.appUrls){
            for (int i = 0; i < config.appUrls.size(); i++) {
                def appUrl = config.appUrls[i]
                if (!(appUrl instanceof Map)) {
                    error "[${STEP_NAME}] The element ${appUrl} is not of type map. Please provide appUrls as a list of maps. For example:\n" +
                            "appUrls: \n" + "  - url: 'https://my-url.com'\n" + "    credentialId: myCreds"
                }
                if (!appUrl.url) {
                    error "[${STEP_NAME}] No url property was defined for the following element in appUrls: ${appUrl}"
                }
                if (appUrl.credentialId) {
                    credentials.add(usernamePassword(credentialsId: appUrl.credentialId, passwordVariable: 'e2e_password', usernameVariable: 'e2e_username'))
                    if (config.wdi5) {
                        credentials.add(usernamePassword(credentialsId: appUrl.credentialId, passwordVariable: 'wdi5_password', usernameVariable: 'wdi5_username'))
                    }
                }
                List scriptOptions = ["--launchUrl=${appUrl.url}"]
                if (appUrl.parameters) {
                    if (appUrl.parameters instanceof List) {
                        scriptOptions = scriptOptions + appUrl.parameters
                    } else {
                        error "[${STEP_NAME}] The parameters property is not of type list. Please provide parameters as a list of strings."
                    }
                }
                e2ETests = runTest(scriptOptions, credentials, e2ETests, index, script, config, npmParameters, stageName)
                index++
            }
        }else{
            if (config.credentialsId) {
                credentials.add(usernamePassword(credentialsId: config.credentialsId, passwordVariable: 'e2e_password', usernameVariable: 'e2e_username'))
                if (config.wdi5) {
                    credentials.add(usernamePassword(credentialsId: config.credentialsId, passwordVariable: 'wdi5_password', usernameVariable: 'wdi5_username'))
                }
            }
            List scriptOptions = []
            if (config.baseUrl){
                scriptOptions = ["--baseUrl=${config.baseUrl}"]
            }
            e2ETests = runTest(scriptOptions, credentials, e2ETests, index, script, config, npmParameters, stageName)
        }
        runClosures(script, e2ETests, config.parallelExecution, "end to end tests")
    }
}

def runTest(scriptOptions, credentials, e2ETests, index, script, config, npmParameters, stageName){
    Closure e2eTest = {
        Utils utils = new Utils()
        utils.unstashStageFiles(script, stageName)
        try {
            withCredentials(credentials) {
                npmExecuteScripts(script: script, parameters: npmParameters, virtualFrameBuffer: true, runScripts: [config.runScript], dockerImage: config.dockerImage, scriptOptions: scriptOptions, buildDescriptorExcludeList: config.buildDescriptorExcludeList)
            }

        } catch (Exception e) {
            error "[${STEP_NAME}] The execution failed with error: ${e.getMessage()}"
        } finally {
            List cucumberFiles = findFiles(glob: "**/e2e/*.json")
            List junitFiles = findFiles(glob: "**/e2e/*.xml")

            if (cucumberFiles.size() > 0) {
                testsPublishResults script: script, cucumber: [active: true, archive: true]
            } else if (junitFiles.size() > 0) {
                testsPublishResults script: script, junit: [active: true, archive: true]
            } else {
                echo "[${STEP_NAME}] No JUnit or cucumber report files found, skipping report visualization."
            }

            utils.stashStageFiles(script, stageName)
        }
    }
    e2ETests["E2E Tests ${index > 1 ? index : ''}"] = {
        if (env.POD_NAME || env.ON_K8S) {
            dockerExecuteOnKubernetes(script: script, containerMap: ContainerMap.instance.getMap().get(stageName) ?: [:]) {
                e2eTest.call()
            }
        } else {
            node(env.NODE_NAME) {
                e2eTest.call()
            }
        }
    }
    return e2ETests
}
