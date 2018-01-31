#!groovy
import com.lesfurets.jenkins.unit.BasePipelineTest

import org.junit.Rule
import org.junit.Test
import static org.junit.Assert.assertTrue
import org.junit.rules.RuleChain

import util.Rules

class DurationMeasureTest extends BasePipelineTest {

    @Rule
    public RuleChain rules = Rules.getCommonRules(this)

    @Test
    void testDurationMeasurement() throws Exception {
        def cpe = loadScript("commonPipelineEnvironment.groovy").commonPipelineEnvironment
        def script = loadScript("durationMeasure.groovy")
        def bodyExecuted = false
        script.call(script: [commonPipelineEnvironment: cpe], measurementName: 'test') {
            bodyExecuted = true
        }
        assertTrue(cpe.getPipelineMeasurement('test') != null)
        assertTrue(bodyExecuted)
        assertJobStatusSuccess()
    }
}
