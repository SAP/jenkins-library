#!groovy
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.hamcrest.Matchers.containsString

import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import static org.junit.Assert.assertThat

import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

class HandlePipelineStepErrorsTest extends BasePiperTest {
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(loggingRule)
        .around(stepRule)
        .around(thrown)

    @Test
    void testBeginAndEndMessage() {
        def isExecuted
        stepRule.step.handlePipelineStepErrors([
            stepName: 'testStep',
            stepParameters: ['something': 'anything']
        ]) {
            isExecuted = true
        }
        // asserts
        assertThat(isExecuted, is(true))
        assertThat(loggingRule.log, containsString('--- Begin library step of: testStep'))
        assertThat(loggingRule.log, containsString('--- End library step of: testStep'))
    }

    @Test
    void testNonVerbose() {
        try {
            stepRule.step.handlePipelineStepErrors([
                stepName: 'testStep',
                stepParameters: ['something': 'anything'],
                echoDetails: false
            ]) {
                throw new Exception('TestError')
            }
        } catch (ignore) {
        } finally {
            // asserts
            assertThat(loggingRule.log, not(containsString('--- Begin library step of: testStep')))
            assertThat(loggingRule.log, not(containsString('--- End library step: testStep')))
            assertThat(loggingRule.log, not(containsString('--- An error occurred in the library step: testStep')))
        }
    }

    @Test
    void testErrorsMessage() {
        def isReported
        try {
            stepRule.step.handlePipelineStepErrors([
                stepName: 'testStep',
                stepParameters: ['something': 'anything']
            ]) {
                throw new Exception('TestError')
            }
        } catch (ignore) {
            isReported = true
        } finally {
            // asserts
            assertThat(isReported, is(true))
            assertThat(loggingRule.log, containsString('--- An error occurred in the library step: testStep'))
            assertThat(loggingRule.log, containsString('[something:anything]'))
        }
    }
}
