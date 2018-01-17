#!groovy
import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import util.JenkinsSetupRule

import static com.lesfurets.jenkins.unit.MethodSignature.method
import static org.junit.Assert.assertTrue
import static org.junit.Assert.assertEquals

class InfluxWriteDataTest extends BasePipelineTest {

    Map fileMap = [:]
    Map stepMap = [:]
    String echoLog = ''

    @Rule
    public JenkinsSetupRule setupRule = new JenkinsSetupRule(this)

    @Before
    void setUp() throws Exception {
        super.setUp()

        helper.registerAllowedMethod('readYaml', [Map.class], { map ->
            return [
                general: [productiveBranch: 'develop'],
                steps : [influxWriteData: [influxServer: 'testInflux']]
            ]
        })

        helper.registerAllowedMethod('writeFile', [Map.class],{m -> fileMap[m.file] = m.text})
        helper.registerAllowedMethod('step', [Map.class],{m -> stepMap = m})
        helper.registerAllowedMethod("echo", [String.class], {s -> echoLog += s})
    }


    @Test
    void testInfluxWriteDataWithDefault() throws Exception {

        def cpe = loadScript('commonPipelineEnvironment.groovy').commonPipelineEnvironment
        cpe.setArtifactVersion('1.2.3')
        def script = loadScript("influxWriteData.groovy")
        script.call(script: [commonPipelineEnvironment: cpe])

        assertTrue(echoLog.contains('Artifact version: 1.2.3'))

        assertEquals('testInflux', stepMap.selectedTarget)
        assertEquals(null, stepMap.customPrefix)
        assertEquals([:], stepMap.customData)
        assertEquals([pipeline_data:[:]], stepMap.customDataMap)

        assertTrue(fileMap.containsKey('jenkins_data.json'))
        assertTrue(fileMap.containsKey('pipeline_data.json'))

        assertJobStatusSuccess()
    }

    @Test
    void testInfluxWriteDataNoInflux() throws Exception {

        //reset stepMap
        stepMap = [:]

        def cpe = loadScript('commonPipelineEnvironment.groovy').commonPipelineEnvironment
        cpe.setArtifactVersion('1.2.3')
        def script = loadScript("influxWriteData.groovy")
        script.call(script: [commonPipelineEnvironment: cpe], influxServer: '')

        assertEquals(0, stepMap.size())

        assertTrue(fileMap.containsKey('jenkins_data.json'))
        assertTrue(fileMap.containsKey('pipeline_data.json'))

        assertJobStatusSuccess()
    }

    @Test
    void testInfluxWriteDataNoArtifactVersion() throws Exception {

        //reset stepMap
        stepMap = [:]
        //reset fileMap
        fileMap = [:]

        def cpe = loadScript('commonPipelineEnvironment.groovy').commonPipelineEnvironment
        def script = loadScript("influxWriteData.groovy")
        script.call(script: [commonPipelineEnvironment: cpe])

        assertEquals(0, stepMap.size())
        assertEquals(0, fileMap.size())

        assertTrue(echoLog.contains('no artifact version available -> exiting writeInflux without writing data'))

        assertJobStatusSuccess()
    }
}
