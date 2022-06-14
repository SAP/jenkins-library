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

class AbapEnvironmentPipelineStageCloneRepositoriesTest extends BasePiperTest {
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
        helper.registerAllowedMethod('strategy', [Map], {m ->
            stepsCalled.add('strategy')
        })
        helper.registerAllowedMethod('cloudFoundryCreateServiceKey', [Map.class], {m -> stepsCalled.add('cloudFoundryCreateServiceKey')})
        helper.registerAllowedMethod('abapEnvironmentPullGitRepo', [Map.class], {m -> stepsCalled.add('abapEnvironmentPullGitRepo')})
        helper.registerAllowedMethod('abapEnvironmentCheckoutBranch', [Map.class], {m -> stepsCalled.add('abapEnvironmentCheckoutBranch')})
        helper.registerAllowedMethod('abapEnvironmentCloneGitRepo', [Map.class], {m -> stepsCalled.add('abapEnvironmentCloneGitRepo')})
        // assertThat(stepsCalled, hasItem('cloudFoundryCreateServiceKey'))
    }

    @Test
    void testAbapEnvironmentPipelineStageCloneRepositoriesPull() {

        nullScript.commonPipelineEnvironment.configuration.runStage = []
        jsr.step.abapEnvironmentPipelineStageCloneRepositories(script: nullScript, strategy: 'Pull',  host: 'abc.com')

        assertThat(stepsCalled, hasItems('abapEnvironmentPullGitRepo'))
        assertThat(stepsCalled, not(hasItem('cloudFoundryCreateServiceKey')))
        assertThat(stepsCalled, not(hasItems('abapEnvironmentCloneGitRepo')))
        assertThat(stepsCalled, not(hasItems('abapEnvironmentCheckoutBranch')))
    }

    @Test
    void testAbapEnvironmentPipelineStageCloneRepositoriesClone() {

        nullScript.commonPipelineEnvironment.configuration.runStage = []
        jsr.step.abapEnvironmentPipelineStageCloneRepositories(script: nullScript, strategy: 'Clone')

        assertThat(stepsCalled, hasItems('abapEnvironmentCloneGitRepo', 'cloudFoundryCreateServiceKey'))
        assertThat(stepsCalled, not(hasItems('abapEnvironmentPullGitRepo', 'abapEnvironmentCheckoutBranch')))
    }

    @Test
    void testAbapEnvironmentPipelineStageCloneRepositoriesCheckoutPull() {

        nullScript.commonPipelineEnvironment.configuration.runStage = []
        jsr.step.abapEnvironmentPipelineStageCloneRepositories(script: nullScript, strategy: 'CheckoutPull')

        assertThat(stepsCalled, hasItems('abapEnvironmentPullGitRepo', 'abapEnvironmentCheckoutBranch', 'cloudFoundryCreateServiceKey'))
        assertThat(stepsCalled, not(hasItems('abapEnvironmentCloneGitRepo')))
    }

    @Test
    void testAbapEnvironmentPipelineStageCloneRepositoriesPullCheckoutPull() {

        nullScript.commonPipelineEnvironment.configuration.runStage = []
        jsr.step.abapEnvironmentPipelineStageCloneRepositories(script: nullScript, strategy: 'AddonBuild', host: 'abc.com')

        assertThat(stepsCalled, not(hasItems('abapEnvironmentPullGitRepo', 'abapEnvironmentCheckoutBranch', 'cloudFoundryCreateServiceKey')))
        assertThat(stepsCalled, hasItems('abapEnvironmentCloneGitRepo'))
    }

    @Test
    void testAbapEnvironmentPipelineStageCloneRepositoriesNoStrategy() {

        nullScript.commonPipelineEnvironment.configuration.runStage = []
        jsr.step.abapEnvironmentPipelineStageCloneRepositories(script: nullScript, host: 'abc.com')

        assertThat(stepsCalled, hasItems('abapEnvironmentPullGitRepo'))
        assertThat(stepsCalled, not(hasItems('abapEnvironmentCloneGitRepo', 'abapEnvironmentCheckoutBranch', 'cloudFoundryCreateServiceKey')))
    }
}
