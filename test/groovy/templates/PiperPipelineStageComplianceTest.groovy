package templates

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertNotNull
import static org.junit.Assert.assertThat

class PiperPipelineStageComplianceTest extends BasePiperTest {
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
        binding.variables.env.STAGE_NAME = 'Compliance'
        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            return body()
        })
        helper.registerAllowedMethod('sonarExecuteScan', [Map.class], {m ->
            stepsCalled.add('sonarExecuteScan')
            stepParameters.sonarExecuteScan = m
        })
    }

    @Test
    void testStageDefault() {
        jsr.step.piperPipelineStageCompliance(
            script: nullScript,
            juStabUtils: utils,
        )
        assertThat(stepsCalled, not(anyOf(hasItems('sonarExecuteScan'))))
    }

    @Test
    void testSonarExecuteScan() {
        jsr.step.piperPipelineStageCompliance(
            script: nullScript,
            juStabUtils: utils,
            sonarExecuteScan: true
        )
        assertThat(stepsCalled, hasItems('sonarExecuteScan'))
        assertNotNull(stepParameters.sonarExecuteScan)
    }
}
