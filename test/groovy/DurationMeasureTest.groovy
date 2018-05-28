#!groovy
import com.lesfurets.jenkins.unit.BasePipelineTest

import org.junit.Rule
import org.junit.Test
import static org.junit.Assert.assertTrue
import org.junit.rules.RuleChain

import util.Rules
import util.JenkinsStepRule
import util.JenkinsEnvironmentRule

class DurationMeasureTest extends BasePipelineTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jsr)
        .around(jer)

    @Test
    void testDurationMeasurement() throws Exception {
        def bodyExecuted = false
        jsr.step.call(script: [commonPipelineEnvironment: jer.env], measurementName: 'test') {
            bodyExecuted = true
        }
        assertTrue(jer.env.getPipelineMeasurement('test') != null)
        assertTrue(bodyExecuted)
        assertJobStatusSuccess()
    }
}
