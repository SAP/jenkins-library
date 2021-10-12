package templates

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.PipelineWhenException
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class AbapEnvironmentPipelineStageATCTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jsr)

    private stepsCalled = []

    @Before
    void init()  {
        binding.variables.env.STAGE_NAME = 'ATC'

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('ATC'))
            return body()
        })
        helper.registerAllowedMethod('abapEnvironmentRunATCCheck', [Map.class], {m -> stepsCalled.add('abapEnvironmentRunATCCheck')})
    }

    @Test
    void testAbapEnvironmentRunTests() {

        nullScript.commonPipelineEnvironment.configuration.runStage = []
        jsr.step.abapEnvironmentPipelineStageATC(script: nullScript)

        assertThat(stepsCalled, hasItems('abapEnvironmentRunATCCheck'))
    }
}
