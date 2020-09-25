import com.sap.piper.DefaultValueCache
import com.sap.piper.JenkinsUtils
import com.sap.piper.analytics.InfluxData
import com.sap.piper.Utils

import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsStepRule
import util.JenkinsReadYamlRule
import util.Rules

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasKey
import static org.hamcrest.Matchers.hasValue
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.isEmptyOrNullString
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue
import static org.junit.Assert.assertEquals

class InfluxWriteDataTest extends BasePiperTest {
    public JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(loggingRule)
        .around(stepRule)

    Map fileMap = [:]
    Map stepMap = [:]
    String echoLog = ''
    String influxVersion

    class JenkinsUtilsMock extends JenkinsUtils {
        String getPluginVersion(name) {
            return influxVersion
        }
    }

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
        influxVersion = '1.15'

        helper.registerAllowedMethod('readYaml', [Map.class], { map ->
            return [
                general: [productiveBranch: 'develop'],
                steps : [influxWriteData: [influxServer: 'testInflux']]
            ]
        })
        helper.registerAllowedMethod('writeFile', [Map.class],{m -> fileMap[m.file] = m.text})
        helper.registerAllowedMethod('step', [Map.class],{m -> stepMap = m})

        helper.registerAllowedMethod('influxDbPublisher', [Map.class],{m -> stepMap = m})

        Utils.metaClass.echo = { def m -> }

    }

    @After
    void teadDown() {
        Utils.metaClass = null
    }

    @Test
    void testInfluxWriteDataWithDefault() throws Exception {

        nullScript.commonPipelineEnvironment.setArtifactVersion('1.2.3')
        stepRule.step.influxWriteData(
            script: nullScript,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
        )

        assertThat(loggingRule.log, containsString('Artifact version: 1.2.3'))

        assertThat(stepMap.selectedTarget, is('testInflux'))
        assertThat(stepMap.customPrefix, isEmptyOrNullString())

        assertThat(stepMap.customData, isEmptyOrNullString())
        assertThat(stepMap.customDataMap, is([pipeline_data: [:], step_data: [:]]))

        assertThat(fileMap, hasKey('jenkins_data.json'))
        assertThat(fileMap, hasKey('influx_data.json'))
        assertThat(fileMap, hasKey('jenkins_data_tags.json'))
        assertThat(fileMap, hasKey('influx_data_tags.json'))

        assertThat(stepMap['$class'], is('InfluxDbPublisher'))

        assertJobStatusSuccess()
    }

    @Test
    void testInfluxWriteDataNoInflux() throws Exception {

        nullScript.commonPipelineEnvironment.setArtifactVersion('1.2.3')
        stepRule.step.influxWriteData(script: nullScript, influxServer: '')

        assertEquals(0, stepMap.size())

        assertTrue(fileMap.containsKey('jenkins_data.json'))
        assertTrue(fileMap.containsKey('influx_data.json'))

        assertJobStatusSuccess()
    }

    @Test
    void testInfluxWriteDataNoArtifactVersion() throws Exception {

        stepRule.step.influxWriteData(script: nullScript)

        assertEquals(0, stepMap.size())
        assertEquals(0, fileMap.size())

        assertTrue(loggingRule.log.contains('no artifact version available -> exiting writeInflux without writing data'))

        assertJobStatusSuccess()
    }

    @Test
    void testInfluxWriteDataWrapInNode() throws Exception {

        boolean nodeCalled = false
        helper.registerAllowedMethod('node', [String.class, Closure.class]) {s, body ->
            nodeCalled = true
            return body()
        }

        helper.registerAllowedMethod("deleteDir", [], null)

        nullScript.commonPipelineEnvironment.setArtifactVersion('1.2.3')
        stepRule.step.influxWriteData(
            script: nullScript,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            wrapInNode: true
        )

        assertThat(nodeCalled, is(true))

    }

    @Test
    void testInfluxCustomData() {
        nullScript.commonPipelineEnvironment.setArtifactVersion('1.2.3')
        stepRule.step.influxWriteData(
            //juStabUtils: utils,
            script: nullScript,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            influxServer: 'myInstance',
            customData: [key1: 'test1'],
            customDataTags: [tag1: 'testTag1'],
            customDataMap: [test_data: [key1: 'keyValue1']],
            customDataMapTags: [test_data: [tag1: 'tagValue1']]
        )
        assertThat(stepMap.customData, allOf(hasKey('key1'), hasValue('test1')))
        assertThat(stepMap.customDataTags, allOf(hasKey('tag1'), hasValue('testTag1')))
        assertThat(stepMap.customDataMap, hasKey('test_data'))
        assertThat(stepMap.customDataMapTags, hasKey('test_data'))
    }

    @Test
    void testInfluxCustomDataFromCPE() {
        nullScript.commonPipelineEnvironment.reset()
        nullScript.commonPipelineEnvironment.setArtifactVersion('1.2.3')
        InfluxData.addTag('jenkins_custom_data', 'tag1', 'testTag1')
        InfluxData.addField('test_data', 'key1', 'keyValue1')
        InfluxData.addTag('test_data', 'tag1', 'tagValue1')
        stepRule.step.influxWriteData(
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            script: nullScript,
            influxServer: 'myInstance'
        )
        assertThat(stepMap.customData, isEmptyOrNullString())
        assertThat(stepMap.customDataTags, allOf(hasKey('tag1'), hasValue('testTag1')))
        assertThat(stepMap.customDataMap, hasKey('test_data'))
        assertThat(stepMap.customDataMapTags, hasKey('test_data'))
    }

    @Test
    void testInfluxWriteDataPluginVersion2() {

        nullScript.commonPipelineEnvironment.setArtifactVersion('1.2.3')
        influxVersion = '2.0'
        stepRule.step.influxWriteData(
            script: nullScript,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
        )

        assertThat(loggingRule.log, containsString('Artifact version: 1.2.3'))

        assertThat(stepMap.selectedTarget, is('testInflux'))
        assertThat(stepMap.customPrefix, isEmptyOrNullString())

        assertThat(stepMap['$class'], isEmptyOrNullString())
    }

}
