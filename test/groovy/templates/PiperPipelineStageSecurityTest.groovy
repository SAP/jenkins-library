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
        binding.variables.env.STAGE_NAME = 'Security'
        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Security'))
            return body()
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
        assertThat(stepsCalled, not(hasItems('whitesourceExecuteScan')))
    }

    @Test
    void testSecurityStageWhiteSource() {

        jsr.step.piperPipelineStageSecurity(
            script: nullScript,
            juStabUtils: utils,
            whitesourceExecuteScan: true
        )

        assertThat(stepsCalled, hasItem('whitesourceExecuteScan'))
    }
}
