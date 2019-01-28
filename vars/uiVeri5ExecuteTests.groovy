import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
import com.sap.piper.GitUtils
import groovy.text.SimpleTemplateEngine
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    'gitSshKeyCredentialsId',
    'verbose'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'dockerEnvVars',
    'dockerImage',
    'dockerWorkspace',
    'failOnError',
    'gitBranch',
    'installCommand',
    'runCommand',
    'seleniumHostAndPort',
    'sidecarEnvVars',
    'sidecarImage',
    'stashContent',
    'testOptions',
    'testRepository',
    'testServerUrl'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

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

        if(!config.dockerEnvVars.TARGET_SERVER_URL)
            config.dockerEnvVars.TARGET_SERVER_URL = config.testServerUrl

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
