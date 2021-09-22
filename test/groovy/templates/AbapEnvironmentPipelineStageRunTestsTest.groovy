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

class abapEnvironmentPipelineStageRunTestsTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jsr)

    private stepsCalled = []

    @Before
    void init()  {
        binding.variables.env.STAGE_NAME = 'Run Tests'

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Run Tests'))
            return body()
        })
        helper.registerAllowedMethod('input', [Map], {m ->
            stepsCalled.add('input')
            return null
        })
        helper.registerAllowedMethod('abapEnvironmentRunAUnitTest', [Map.class], {m -> stepsCalled.add('abapEnvironmentRunAUnitTest')})
        helper.registerAllowedMethod('abapEnvironmentRunATCCheck', [Map.class], {m -> stepsCalled.add('abapEnvironmentRunATCCheck')})
    }

    @Test
    void testAbapEnvironmentRunTests() {

        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Run Tests': true
        ]
        jsr.step.abapEnvironmentPipelineStageRunTests(script: nullScript)

        assertThat(stepsCalled, hasItems('input'))
        assertThat(stepsCalled, hasItems('abapEnvironmentRunAUnitTest'))
        assertThat(stepsCalled, hasItems('abapEnvironmentRunATCCheck'))
    }
}
