package templates

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.hasItems
import static org.hamcrest.Matchers.not
import static org.junit.Assert.assertEquals
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
            sonarExecuteScan: true
        )
        assertThat(stepsCalled, hasItems('sonarExecuteScan'))
        assertNotNull(stepParameters.sonarExecuteScan)
        assertThat(stepParameters.sonarExecuteScan, not(hasItems('branchName')))
    }

    @Test
    void testStageNonPullRequestBranch() {
        binding.variables.env.BRANCH_NAME = 'test-branch'
        jsr.step.piperPipelineStageCompliance(
            script: nullScript,
            juStabUtils: utils,
            sonarExecuteScan: true
        )
        assertThat(stepsCalled, hasItems('sonarExecuteScan'))
        assertNotNull(stepParameters.sonarExecuteScan)
        assertEquals('test-branch', stepParameters.sonarExecuteScan.branchName)
    }
}
