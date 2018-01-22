#!groovy
import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.JenkinsLoggingRule
import util.JenkinsSetupRule

import static org.junit.Assert.assertTrue
import static org.junit.Assert.assertEquals

class InfluxWriteDataTest extends BasePipelineTest {

    Script influxWriteDataScript

    Map fileMap = [:]
    Map stepMap = [:]
    String echoLog = ''

    def cpe

    public JenkinsSetupRule setupRule = new JenkinsSetupRule(this)
    public JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain =
        RuleChain.outerRule(setupRule)
            .around(loggingRule)

    @Before
    void init() throws Exception {

        //reset stepMap
        stepMap = [:]
        //reset fileMap
        fileMap = [:]

        helper.registerAllowedMethod('readYaml', [Map.class], { map ->
            return [
                general: [productiveBranch: 'develop'],
                steps : [influxWriteData: [influxServer: 'testInflux']]
            ]
        })

        helper.registerAllowedMethod('writeFile', [Map.class],{m -> fileMap[m.file] = m.text})
        helper.registerAllowedMethod('step', [Map.class],{m -> stepMap = m})

        cpe = loadScript('commonPipelineEnvironment.groovy').commonPipelineEnvironment
        influxWriteDataScript = loadScript("influxWriteData.groovy")
    }


    @Test
    void testInfluxWriteDataWithDefault() throws Exception {

        cpe.setArtifactVersion('1.2.3')
        influxWriteDataScript.call(script: [commonPipelineEnvironment: cpe])

        assertTrue(loggingRule.log.contains('Artifact version: 1.2.3'))

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

        cpe.setArtifactVersion('1.2.3')
        influxWriteDataScript.call(script: [commonPipelineEnvironment: cpe], influxServer: '')

        assertEquals(0, stepMap.size())

        assertTrue(fileMap.containsKey('jenkins_data.json'))
        assertTrue(fileMap.containsKey('pipeline_data.json'))

        assertJobStatusSuccess()
    }

    @Test
    void testInfluxWriteDataNoArtifactVersion() throws Exception {

        influxWriteDataScript.call(script: [commonPipelineEnvironment: cpe])

        assertEquals(0, stepMap.size())
        assertEquals(0, fileMap.size())

        assertTrue(loggingRule.log.contains('no artifact version available -> exiting writeInflux without writing data'))

        assertJobStatusSuccess()
    }
}
