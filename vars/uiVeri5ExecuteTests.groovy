import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.GitUtils
import com.sap.piper.Utils

import groovy.text.SimpleTemplateEngine
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * In case a `testRepository` is provided and it is protected, access credentials (as Jenkins credentials) can be provided with `gitSshKeyCredentialsId`. **Note: In case of using a protected repository, `testRepository` should include the ssh link to the repository.**
     * @possibleValues Jenkins credentialId
     */
    'gitSshKeyCredentialsId'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * A map of environment variables to set in the container, e.g. [http_proxy:'proxy:8080'].
     */
    'dockerEnvVars',
    /**
     * The name of the docker image that should be used. If empty, Docker is not used and the command is executed directly on the Jenkins system.
     */
    'dockerImage',
    /**
     * Only relevant for Kubernetes case: Specifies a dedicated user home directory for the container which will be passed as value for environment variable `HOME`.
     */
    'dockerWorkspace',
    /**
     * With `failOnError` the behavior in case tests fail can be defined.
     * @possibleValues `true`, `false`
     */
    'failOnError',
    /**
     * In case a `testRepository` is provided the branch in this repository can be specified with `gitBranch`.
     */
    'gitBranch',
    /**
     * The command that is executed to install the test tool.
     */
    'installCommand',
    /**
     * The command that is executed to start the tests.
     */
    'runCommand',
    /**
     * The host of the selenium hub, this is set automatically to `localhost` in a Kubernetes environment (determined by the `ON_K8S` environment variable) of to `selenium` in any other case. The value is only needed for the `runCommand`.
     */
    'seleniumHost',
    /**
     * The port of the selenium hub. The value is only needed for the `runCommand`.
     */
    'seleniumPort',
    /**
     * A map of environment variables to set in the sidecar container, similar to `dockerEnvVars`.
     */
    'sidecarEnvVars',
    /**
     * The name of the docker image of the sidecar container. If empty, no sidecar container is started.
     */
    'sidecarImage',
    /**
     * If specific stashes should be considered for the tests, their names need to be passed via the parameter `stashContent`.
     */
    'stashContent',
    /**
     * This allows to set specific options for the UIVeri5 execution. Details can be found [in the UIVeri5 documentation](https://github.com/SAP/ui5-uiveri5/blob/master/docs/config/config.md#configuration).
     */
    'testOptions',
    /**
     * With `testRepository` the tests can be loaded from another reposirory.
     */
    'testRepository'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * With this step [UIVeri5](https://github.com/SAP/ui5-uiveri5) tests can be executed.
 *
 * UIVeri5 describes following benefits on its GitHub page:
 *
 * * Automatic synchronization with UI5 app rendering so there is no need to add waits and sleeps to your test. Tests are reliable by design.
 * * Tests are written in synchronous manner, no callbacks, no promise chaining so are really simple to write and maintain.
 * * Full power of webdriverjs, protractor and jasmine - deferred selectors, custom matchers, custom locators.
 * * Control locators (OPA5 declarative matchers) allow locating and interacting with UI5 controls.
 * * Does not depend on testability support in applications - works with autorefreshing views, resizing elements, animated transitions.
 * * Declarative authentications - authentication flow over OAuth2 providers, etc.
 * * Console operation, CI ready, fully configurable, no need for java (comming soon) or IDE.
 * * Covers full ui5 browser matrix - Chrome,Firefox,IE,Edge,Safari,iOS,Android.
 * * Open-source, modify to suite your specific neeeds.
 *
 * !!! note "Browser Matrix"
 *     With this step and the underlying Docker image ([selenium/standalone-chrome](https://github.com/SeleniumHQ/docker-selenium/tree/master/StandaloneChrome)) only Chrome tests are possible.
 *
 *     Testing of further browsers can be done with using a custom Docker image.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .addIfEmpty('seleniumHost', isKubernetes()?'localhost':'selenium')
            .use()

        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], config)

        config.stashContent = config.testRepository ? [GitUtils.handleTestRepository(this, config)] : utils.unstashAll(config.stashContent)
        config.installCommand = SimpleTemplateEngine.newInstance().createTemplate(config.installCommand).make([config: config]).toString()
        config.runCommand = SimpleTemplateEngine.newInstance().createTemplate(config.runCommand).make([config: config]).toString()

        seleniumExecuteTests(
            script: script,
            buildTool: 'npm',
            dockerEnvVars: config.dockerEnvVars,
            dockerImage: config.dockerImage,
            dockerName: config.dockerName,
            dockerWorkspace: config.dockerWorkspace,
            sidecarEnvVars: config.sidecarEnvVars,
            sidecarImage: config.sidecarImage,
            stashContent: config.stashContent
        ) {
            try {
                sh "NPM_CONFIG_PREFIX=~/.npm-global ${config.installCommand}"
                sh "PATH=\$PATH:~/.npm-global/bin ${config.runCommand} ${config.testOptions}"
            } catch (err) {
                echo "[${STEP_NAME}] Test execution failed"
                script.currentBuild.result = 'UNSTABLE'
                if (config.failOnError) throw err
            }
        }
    }
}

boolean isKubernetes() {
    return Boolean.valueOf(env.ON_K8S)
}
