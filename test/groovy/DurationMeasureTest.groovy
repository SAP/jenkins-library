#!groovy
import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.Rule
import org.junit.Test
import util.JenkinsSetupRule
import static org.junit.Assert.assertTrue

class DurationMeasureTest extends BasePipelineTest {

    @Rule
    public JenkinsSetupRule setupRule = new JenkinsSetupRule(this)

    @Test
    void testDurationMeasurement() throws Exception {
        //def cpe = new CPEMock()
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
