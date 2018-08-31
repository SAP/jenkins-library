#!groovy

import org.junit.Rule
import org.junit.Test
import util.BasePiperTest

import static org.junit.Assert.assertTrue
import org.junit.rules.RuleChain

import util.Rules
import util.JenkinsReadYamlRule
import util.JenkinsStepRule


class DurationMeasureTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jsr)

    @Test
    void testDurationMeasurement() throws Exception {
        def bodyExecuted = false
        jsr.step.call(script: nullScript, measurementName: 'test') {
            bodyExecuted = true
        }
        assertTrue(nullScript.commonPipelineEnvironment.getPipelineMeasurement('test') != null)
        assertTrue(bodyExecuted)
        assertJobStatusSuccess()
    }
}
