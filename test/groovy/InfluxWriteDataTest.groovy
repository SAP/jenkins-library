#!groovy
import com.lesfurets.jenkins.unit.BasePipelineTest
import com.sap.piper.DefaultValueCache
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.JenkinsLoggingRule
import util.JenkinsStepRule
import util.JenkinsEnvironmentRule
import util.Rules

import static org.junit.Assert.assertTrue
import static org.junit.Assert.assertEquals

class InfluxWriteDataTest extends BasePipelineTest {
    public JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(loggingRule)
        .around(jsr)
        .around(jer)

    Map fileMap = [:]
    Map stepMap = [:]
    String echoLog = ''

    @Before
    void init() throws Exception {
        // Currently we have dependencies between the tests since
        // DefaultValueCache is a singleton which keeps its status
        // for all the tests. Depending on the test order we fail.
        // As long as this status remains we need:
        DefaultValueCache.reset()
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
    }


    @Test
    void testInfluxWriteDataWithDefault() throws Exception {

        jer.env.setArtifactVersion('1.2.3')
        jsr.step.call(script: [commonPipelineEnvironment: jer.env])

        assertTrue(loggingRule.log.contains('Artifact version: 1.2.3'))

        assertEquals('testInflux', stepMap.selectedTarget)
        assertEquals(null, stepMap.customPrefix)
        assertEquals([:], stepMap.customData)
        assertEquals([pipeline_data: [:], step_data: [:]], stepMap.customDataMap)

        assertTrue(fileMap.containsKey('jenkins_data.json'))
        assertTrue(fileMap.containsKey('pipeline_data.json'))

        assertJobStatusSuccess()
    }

    @Test
    void testInfluxWriteDataNoInflux() throws Exception {

        jer.env.setArtifactVersion('1.2.3')
        jsr.step.call(script: [commonPipelineEnvironment: jer.env], influxServer: '')

        assertEquals(0, stepMap.size())

        assertTrue(fileMap.containsKey('jenkins_data.json'))
        assertTrue(fileMap.containsKey('pipeline_data.json'))

        assertJobStatusSuccess()
    }

    @Test
    void testInfluxWriteDataNoArtifactVersion() throws Exception {

        jsr.step.call(script: [commonPipelineEnvironment: jer.env])

        assertEquals(0, stepMap.size())
        assertEquals(0, fileMap.size())

        assertTrue(loggingRule.log.contains('no artifact version available -> exiting writeInflux without writing data'))

        assertJobStatusSuccess()
    }
}
