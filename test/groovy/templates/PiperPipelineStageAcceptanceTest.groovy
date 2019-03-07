#!groovy
package templates

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.containsString
import static org.junit.Assert.assertThat

class PiperPipelineStageAcceptanceTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jlr)
        .around(jsr)

    @Before
    void init()  {
        binding.variables.env.STAGE_NAME = 'Acceptance'
        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            return body()
        })
    }

    @Test
    void testStageDefault() {

        jsr.step.piperPipelineStageIntegration(
            script: nullScript,
            juStabUtils: utils,
        )
        assertThat(jlr.log, containsString('Stage implementation is not provided yet.'))

    }
}
