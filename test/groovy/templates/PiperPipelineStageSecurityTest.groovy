package templates

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.hasItems
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.junit.Assert.assertThat

class PiperPipelineStageSecurityTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jlr)
        .around(jsr)

    private List stepsCalled = []
    private Map stepParameters = [:]

    @Before
    void init()  {
        nullScript.env.STAGE_NAME = 'Security'

        helper.registerAllowedMethod("deleteDir", [], null)
        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Security'))
            return body()
        })

        def parallelMap = [:]
        helper.registerAllowedMethod("parallel", [Map.class], { map ->
            parallelMap = map
            parallelMap.each {key, value ->
                if (key != 'failFast') {
                    value()
                }
            }
        })

        helper.registerAllowedMethod('checkmarxExecuteScan', [Map.class], {m ->
            stepsCalled.add('checkmarxExecuteScan')
        })

        helper.registerAllowedMethod('detectExecuteScan', [Map.class], {m ->
            stepsCalled.add('detectExecuteScan')
        })

        helper.registerAllowedMethod('fortifyExecuteScan', [Map.class], {m ->
            stepsCalled.add('fortifyExecuteScan')
        })

        helper.registerAllowedMethod('whitesourceExecuteScan', [Map.class], {m ->
            stepsCalled.add('whitesourceExecuteScan')
            stepParameters.whitesourceExecuteScan = m
        })
    }

    @Test
    void testStageDefault() {

        jsr.step.piperPipelineStageSecurity(
            script: nullScript,
            juStabUtils: utils,
        )
        assertThat(stepsCalled, not(hasItem('whitesourceExecuteScan')))
        assertThat(stepsCalled, not(hasItem('checkmarxExecuteScan')))
        assertThat(stepsCalled, not(hasItem('detectExecuteScan')))
        assertThat(stepsCalled, not(hasItem('fortifyExecuteScan')))
    }

    @Test
    void testSecurityStageWhiteSource() {

        jsr.step.piperPipelineStageSecurity(
            script: nullScript,
            juStabUtils: utils,
            whitesourceExecuteScan: true
        )

        assertThat(stepsCalled, hasItem('whitesourceExecuteScan'))
        assertThat(stepsCalled, not(hasItem('checkmarxExecuteScan')))
        assertThat(stepsCalled, not(hasItem('detectExecuteScan')))
        assertThat(stepsCalled, not(hasItem('fortifyExecuteScan')))
    }

    @Test
    void testSecurityStageCheckmarx() {

        nullScript.commonPipelineEnvironment.configuration = [
            runStep: [Security: [checkmarxExecuteScan: true]]
        ]

        jsr.step.piperPipelineStageSecurity(
            script: nullScript,
            juStabUtils: utils,
            checkmarxExecuteScan: true
        )

        assertThat(stepsCalled, hasItem('checkmarxExecuteScan'))
        assertThat(stepsCalled, not(hasItem('detectExecuteScan')))
        assertThat(stepsCalled, not(hasItem('whitesourceExecuteScan')))
        assertThat(stepsCalled, not(hasItem('fortifyExecuteScan')))
    }

    @Test
    void testSecurityStageDetect() {

        nullScript.commonPipelineEnvironment.configuration = [
            runStep: [Security: [detectExecuteScan: true]]
        ]

        jsr.step.piperPipelineStageSecurity(
            script: nullScript,
            juStabUtils: utils,
            detectExecuteScan: true
        )

        assertThat(stepsCalled, hasItem('detectExecuteScan'))
        assertThat(stepsCalled, not(hasItem('checkmarxExecuteScan')))
        assertThat(stepsCalled, not(hasItem('whitesourceExecuteScan')))
        assertThat(stepsCalled, not(hasItem('fortifyExecuteScan')))
    }

    @Test
    void testSecurityStageFortify() {

        nullScript.commonPipelineEnvironment.configuration = [
            runStep: [Security: [fortifyExecuteScan: true]]
        ]

        jsr.step.piperPipelineStageSecurity(
            script: nullScript,
            juStabUtils: utils,
            fortifyExecuteScan: true
        )

        assertThat(stepsCalled, hasItem('fortifyExecuteScan'))
        assertThat(stepsCalled, not(hasItem('detectExecuteScan')))
        assertThat(stepsCalled, not(hasItem('whitesourceExecuteScan')))
        assertThat(stepsCalled, not(hasItem('checkmarxExecuteScan')))
    }
}
