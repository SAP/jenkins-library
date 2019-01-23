import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
import com.sap.piper.GitUtils
import groovy.text.SimpleTemplateEngine
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field Set STEP_CONFIG_KEYS = [
    'buildTool',
    'dockerEnvVars',
    'dockerImage',
    'dockerName',
    'dockerWorkspace',
    'failOnError',
    'gitBranch',
    'gitSshKeyCredentialsId',
    'installCommand',
    'languageRunner',
    'runCommand',
    'stashContent',
    'testOptions',
    'testRepository',
    'testServerUrl'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters)  ?: this
        def utils = parameters.juStabUtils ?: new Utils()

        script.commonPipelineEnvironment.setInfluxStepData('gauge', false)

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .dependingOn('buildTool').mixin('dockerImage')
            .dependingOn('buildTool').mixin('dockerName')
            .dependingOn('buildTool').mixin('dockerWorkspace')
            .dependingOn('buildTool').mixin('languageRunner')
            .dependingOn('buildTool').mixin('runCommand')
            .dependingOn('buildTool').mixin('testOptions')
            .use()

        utils.pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'buildTool',
            stepParam1: config.buildTool,
            stepParamKey2: 'dockerName',
            stepParam2: config.dockerName
        ], config)

        if(!config.dockerEnvVars.TARGET_SERVER_URL && config.testServerUrl)
            config.dockerEnvVars.TARGET_SERVER_URL = config.testServerUrl

        if (config.testRepository) {
            // handle separate test repository
            config.stashContent = [GitUtils.handleTestRepository(this, config)]
        } else {
            config.stashContent = utils.unstashAll(config.stashContent)
        }

        seleniumExecuteTests (
            script: script,
            buildTool: config.buildTool,
            dockerEnvVars: config.dockerEnvVars,
            dockerImage: config.dockerImage,
            dockerName: config.dockerName,
            dockerWorkspace: config.dockerWorkspace,
            stashContent: config.stashContent
        ) {
            String gaugeScript = ''
            if (config.installCommand) {
                gaugeScript = '''export HOME=${HOME:-$(pwd)}
                    if [ "$HOME" = "/" ]; then export HOME=$(pwd); fi
                    export PATH=$HOME/bin/gauge:$PATH
                    mkdir -p $HOME/bin/gauge
                    ''' + config.installCommand + '''
                    gauge telemetry off
                    gauge install ''' + config.languageRunner + '''
                    gauge install html-report
                    gauge install xml-report
                    '''
            }
            gaugeScript += config.runCommand

            try {
                sh "${gaugeScript} ${config.testOptions}"
                script.commonPipelineEnvironment.setInfluxStepData('gauge', true)
            } catch (err) {
                echo "[${STEP_NAME}] One or more tests failed"
                script.currentBuild.result = 'UNSTABLE'
                if (config.failOnError) throw err
            }
        }
    }
}
