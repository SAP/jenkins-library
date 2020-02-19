import com.sap.piper.DebugReport
import hudson.AbortException

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
            assertThat(loggingRule.log, containsString('to show step parameters, set verbose:true'))
        }
    }

    @Test
    void testHandleErrorsIgnoreFailure() {
        def errorOccured = false
        helper.registerAllowedMethod('unstable', [String.class], {s ->
            nullScript.currentBuild.result = 'UNSTABLE'
        })
        try {
            stepRule.step.handlePipelineStepErrors([
                stepName: 'test',
                stepParameters: [jenkinsUtilsStub: jenkinsUtils, script: nullScript],
                failOnError: false
            ]) {
                throw new AbortException('TestError')
            }
        } catch (err) {
            errorOccured = true
        }
        assertThat(errorOccured, is(false))
        assertThat(nullScript.currentBuild.result, is('UNSTABLE'))
    }

    @Test
    void testHandleErrorsIgnoreFailureBlacklist() {
        def errorOccured = false

        //define blacklist in defaults
        helper.registerAllowedMethod("readYaml", [Map], { Map m ->
            return [steps: [handlePipelineStepErrors: [mandatorySteps: ['step1', 'test']]]]
        })

        try {
            stepRule.step.handlePipelineStepErrors([
                stepName: 'test',
                stepParameters: [jenkinsUtilsStub: jenkinsUtils, script: nullScript],
                failOnError: false
            ]) {
                throw new AbortException('TestError')
            }
        } catch (err) {
            errorOccured = true
        }
        assertThat(errorOccured, is(true))
    }

    @Test
    void testHandleErrorsIgnoreFailureNoScript() {
        def errorOccured = false
        helper.registerAllowedMethod('unstable', [String.class], {s ->
            //test behavior in case plugina are not yet up to date
            throw new java.lang.NoSuchMethodError('No such DSL method \'unstable\' found')
        })
        try {
            stepRule.step.handlePipelineStepErrors([
                stepName: 'test',
                stepParameters: [jenkinsUtilsStub: jenkinsUtils],
                failOnError: false
            ]) {
                throw new AbortException('TestError')
            }
        } catch (err) {
            errorOccured = true
        }
        assertThat(errorOccured, is(false))
    }

    @Test
    void testHandleErrorsTimeout() {
        def timeout = 0
        helper.registerAllowedMethod('timeout', [Map.class, Closure.class], {m, body ->
            timeout = m.time
            throw new org.jenkinsci.plugins.workflow.steps.FlowInterruptedException(hudson.model.Result.ABORTED, new jenkins.model.CauseOfInterruption.UserInterruption('Test'))
        })
        String errorMsg
        helper.registerAllowedMethod('unstable', [String.class], {s ->
            nullScript.currentBuild.result = 'UNSTABLE'
            errorMsg = s
        })

        stepRule.step.handlePipelineStepErrors([
            stepName: 'test',
            stepParameters: [jenkinsUtilsStub: jenkinsUtils, script: nullScript],
            failOnError: false,
            stepTimeouts: [test: 10]
        ]) {
            //do something
        }
        assertThat(timeout, is(10))
        assertThat(nullScript.currentBuild.result, is('UNSTABLE'))
        assertThat(errorMsg, is('[handlePipelineStepErrors] Error in step test - Build result set to \'UNSTABLE\''))
    }

    @Test
    void testFeedDebugReport() {
        Exception err = new Exception('TestError')
        try {
            stepRule.step.handlePipelineStepErrors([
                stepName: 'testStep',
                stepParameters: ['something': 'anything'],
            ]) {
                throw err
            }
        } catch (ignore) {
        } finally {
            // asserts
            assertThat(DebugReport.instance.failedBuild.step, is('testStep'))
            assertThat(DebugReport.instance.failedBuild.fatal, is('true'))
            assertThat(DebugReport.instance.failedBuild.reason, is(err))
        }
    }

}
