import org.jenkinsci.plugins.workflow.steps.FlowInterruptedException
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.CoreMatchers.containsString
import static org.hamcrest.CoreMatchers.is
import static org.junit.Assert.assertThat

class PipelineRestartStepsTest extends BasePiperTest {

    private JenkinsErrorRule errorRule = new JenkinsErrorRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)

    @Rule
    public RuleChain chain = Rules.getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(errorRule)
        .around(loggingRule)
        .around(stepRule)

    @Test
    void testError() throws Exception {

        def mailBuildResult = ''
        helper.registerAllowedMethod('mailSendNotification', [Map.class], { m ->
            mailBuildResult = m.buildResult
            return null
        })

        helper.registerAllowedMethod('timeout', [Map.class, Closure.class], { m, closure ->
            assertThat(m.time, is(1))
            assertThat(m.unit, is('SECONDS'))
            return closure()
        })

        def iterations = 0
        helper.registerAllowedMethod('input', [Map.class], { m ->
            iterations ++
            assertThat(m.message, is('Do you want to restart?'))
            assertThat(m.ok, is('Restart'))
            if (iterations > 1) {
                throw new FlowInterruptedException()
            } else {
                return null
            }
        })

        try {
            stepRule.step.pipelineRestartSteps ([
                script: nullScript,
                jenkinsUtilsStub: jenkinsUtils,
                sendMail: true,
                timeoutInSeconds: 1

            ]) {
                throw new hudson.AbortException('I just created an error')
            }
        } catch(err) {
            assertThat(loggingRule.log, containsString('ERROR occurred: hudson.AbortException: I just created an error'))
            assertThat(mailBuildResult, is('UNSTABLE'))
        }
    }

    @Test
    void testErrorNoMail() throws Exception {

        def mailBuildResult = ''
        helper.registerAllowedMethod('mailSendNotification', [Map.class], { m ->
            mailBuildResult = m.buildResult
            return null
        })

        helper.registerAllowedMethod('timeout', [Map.class, Closure.class], { m, closure ->
            assertThat(m.time, is(1))
            assertThat(m.unit, is('SECONDS'))
            return closure()
        })

        def iterations = 0
        helper.registerAllowedMethod('input', [Map.class], { m ->
            iterations ++
            assertThat(m.message, is('Do you want to restart?'))
            assertThat(m.ok, is('Restart'))
            if (iterations > 1) {
                throw new FlowInterruptedException()
            } else {
                return null
            }
        })

        try {
            stepRule.step.pipelineRestartSteps ([
                script: nullScript,
                jenkinsUtilsStub: jenkinsUtils,
                sendMail: false,
                timeoutInSeconds: 1

            ]) {
                throw new hudson.AbortException('I just created an error')
            }
        } catch(err) {
            assertThat(loggingRule.log, containsString('ERROR occurred: hudson.AbortException: I just created an error'))
            assertThat(mailBuildResult, is(''))
        }
    }

    @Test
    void testSuccess() throws Exception {

        stepRule.step.pipelineRestartSteps ([
            script: nullScript,
            jenkinsUtilsStub: jenkinsUtils,
            sendMail: false,
            timeoutInSeconds: 1

        ]) {
            nullScript.echo 'This is a test'
        }

        assertThat(loggingRule.log, containsString('This is a test'))
    }
}
