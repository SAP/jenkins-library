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
import util.JenkinsStepRule
import util.Rules

class HandlePipelineStepErrorsTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jsr)
        .around(thrown)

    @Test
    void testBeginAndEndMessage() {
        def isExecuted
        jsr.step.handlePipelineStepErrors([
            stepName: 'testStep',
            stepParameters: ['something': 'anything']
        ]) {
            isExecuted = true
        }
        // asserts
        assertThat(isExecuted, is(true))
        assertThat(jlr.log, containsString('--- BEGIN LIBRARY STEP: testStep'))
        assertThat(jlr.log, containsString('--- END LIBRARY STEP: testStep'))
    }

    @Test
    void testNonVerbose() {
        try {
            jsr.step.handlePipelineStepErrors([
                stepName: 'testStep',
                stepParameters: ['something': 'anything'],
                echoDetails: false
            ]) {
                throw new Exception('TestError')
            }
        } catch (ignore) {
        } finally {
            // asserts
            assertThat(jlr.log, not(containsString('--- BEGIN LIBRARY STEP: testStep')))
            assertThat(jlr.log, not(containsString('--- END LIBRARY STEP: testStep')))
            assertThat(jlr.log, not(containsString('--- ERROR OCCURRED IN LIBRARY STEP: testStep')))
        }
    }

    @Test
    void testErrorsMessage() {
        def isReported
        try {
            jsr.step.handlePipelineStepErrors([
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
            assertThat(jlr.log, containsString('--- ERROR OCCURRED IN LIBRARY STEP: testStep'))
            assertThat(jlr.log, containsString('[something:anything]'))
        }
    }
}
