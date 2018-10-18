import com.sap.piper.JenkinsUtils
import com.sap.piper.ConfigurationHelper
import groovy.transform.Field

@Field String STEP_NAME = 'pipelineRestartSteps'
@Field Set STEP_CONFIG_KEYS = [
    'sendMail',
    'timeoutInSeconds'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:], body) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        def script = parameters.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()
        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        def restart = true
        while (restart) {
            try {
                body()
                restart = false
            } catch (Throwable err) {
                echo "ERROR occured: ${err}"
                if (config.sendMail)
                    if (jenkinsUtils.nodeAvailable()) {
                        mailSendNotification script: script, buildResult: 'UNSTABLE'
                    } else {
                        node {
                            mailSendNotification script: script, buildResult: 'UNSTABLE'
                        }
                    }

                try {
                    timeout(time: config.timeoutInSeconds, unit: 'SECONDS') {
                        input message: 'Do you want to restart?', ok: 'Restart'
                    }
                } catch(e) {
                    restart = false
                    throw err
                }
            }
        }
    }
}
