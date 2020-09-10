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

class AbapEnvironmentPipelineStagePrepareSystemTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jsr)

    private stepsCalled = []

    @Before
    void init()  {
        binding.variables.env.STAGE_NAME = 'Clone Repositories'

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Clone Repositories'))
            return body()
        })
        helper.registerAllowedMethod('input', [Map], {m -> return null})
        helper.registerAllowedMethod('abapEnvironmentPullGitRepo', [Map.class], {m -> stepsCalled.add('abapEnvironmentPullGitRepo')})
        helper.registerAllowedMethod('abapEnvironmentCheckoutBranch', [Map.class], {m -> stepsCalled.add('abapEnvironmentCheckoutBranch')})
    }

    @Test
    void testAbapEnvironmentPipelineStageCloneRepositoriesExecuted() {

        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Prepare System': false
        ]
        jsr.step.abapEnvironmentPipelineStagePrepareSystem(script: nullScript)

        assertThat(stepsCalled, hasItems('abapEnvironmentPullGitRepo', 'abapEnvironmentCheckoutBranch'))
    }

    @Test
    void testAbapEnvironmentPipelineStageCloneRepositoriesNotExecuted() {

        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Prepare System': true
        ]
        jsr.step.abapEnvironmentPipelineStagePrepareSystem(script: nullScript)

        assertThat(stepsCalled, not(hasItems('abapEnvironmentPullGitRepo', 'abapEnvironmentCheckoutBranch')))
    }
}
