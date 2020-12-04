import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import com.sap.piper.JenkinsUtils
import com.sap.piper.ConfigurationHelper
import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []

@Field Set STEP_CONFIG_KEYS = [
    /**
     * If it is set to `true` the step `mailSendNotification` will be triggered in case of an error.
     */
    'sendMail',
    /**
     *  If it is set, the step message can be customized to throw user friendly error messages in Jenkins UI.
     */
    'stepMessage',
    /**
     * Defines the time period where the job waits for input. Default is 15 minutes. Once this time is passed the job enters state `FAILED`.
     */
    'timeoutInSeconds'
]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Support of restarting failed stages or steps in a pipeline is limited in Jenkins.
 *
 * This has been documented in the [Jenkins Jira issue JENKINS-33846](https://issues.jenkins-ci.org/browse/JENKINS-33846).
 *
 * For declarative pipelines there is a solution available which partially addresses this topic:
 * https://jenkins.io/doc/book/pipeline/running-pipelines/#restart-from-a-stage.
 *
 * Nonetheless, still features are missing, so it can't be used in all cases.
 * The more complex Piper pipelines which share a state via [`commonPipelineEnvironment`](commonPipelineEnvironment.md) will for example not work with the standard _restart-from-stage_.
 *
 * The step `pipelineRestartSteps` aims to address this gap and allows individual parts of a pipeline (e.g. a failed deployment) to be restarted.
 *
 * This is done in a way that the pipeline waits for user input to restart the pipeline in case of a failure. In case this user input is not provided the pipeline stops after a timeout which can be configured.
 */
@GenerateDocumentation
void call(Map parameters = [:], body) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, failOnError: true) {
        def script = checkScript(this, parameters) ?: this
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()
        String stageName = parameters.stageName ?: env.STAGE_NAME

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        def restart = true
        while (restart) {
            try {
                body()
                restart = false
            } catch (Throwable err) {
                echo "ERROR occurred: ${err}"
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
                        input message: config.stepMessage, ok: 'Restart'
                    }
                } catch(e) {
                    restart = false
                    throw err
                }
            }
        }
    }
}
