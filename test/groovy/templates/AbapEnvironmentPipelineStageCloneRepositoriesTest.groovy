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
        helper.registerAllowedMethod('abapEnvironmentPullGitRepo', [Map.class], {m -> stepsCalled.add('abapEnvironmentPullGitRepo')})
        helper.registerAllowedMethod('abapEnvironmentCheckoutBranch', [Map.class], {m -> stepsCalled.add('abapEnvironmentCheckoutBranch')})
        helper.registerAllowedMethod('abapEnvironmentCloneGitRepo', [Map.class], {m -> stepsCalled.add('abapEnvironmentCloneGitRepo')})
    }
    
    @Test
    void testAbapEnvironmentPipelineStageCloneRepositoriesPull() {

        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Prepare System': true
        ]
        jsr.step.abapEnvironmentPipelineStageCloneRepositories(script: nullScript, strategy: 'Pull')

        assertThat(stepsCalled, hasItems('abapEnvironmentPullGitRepo'))
        assertThat(stepsCalled, not(hasItems('abapEnvironmentCloneGitRepo')))
        assertThat(stepsCalled, not(hasItems('abapEnvironmentCheckoutBranch')))
    }

    @Test
    void testAbapEnvironmentPipelineStageCloneRepositoriesClone() {

        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Prepare System': true
        ]
        
        jsr.step.abapEnvironmentPipelineStageCloneRepositories(script: nullScript, strategy: 'Clone')

        assertThat(stepsCalled, hasItems('abapEnvironmentCloneGitRepo'))
        assertThat(stepsCalled, not(hasItems('abapEnvironmentPullGitRepo', 'abapEnvironmentCheckoutBranch')))
    }

    @Test
    void testAbapEnvironmentPipelineStageCloneRepositoriesCheckoutPull() {

        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Prepare System': true
        ]
        
        jsr.step.abapEnvironmentPipelineStageCloneRepositories(script: nullScript, strategy: 'CheckoutPull')

        assertThat(stepsCalled, hasItems('abapEnvironmentPullGitRepo', 'abapEnvironmentCheckoutBranch'))
        assertThat(stepsCalled, not(hasItems('abapEnvironmentCloneGitRepo')))
    }

    @Test
    void testAbapEnvironmentPipelineStageCloneRepositoriesPullCheckoutPull() {

        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Prepare System': true
        ]
        
        jsr.step.abapEnvironmentPipelineStageCloneRepositories(script: nullScript, strategy: 'addonBuild')

        assertThat(stepsCalled, hasItems('abapEnvironmentPullGitRepo', 'abapEnvironmentCheckoutBranch'))
        assertThat(stepsCalled, not(hasItems('abapEnvironmentCloneGitRepo')))
    }

    @Test
    void testAbapEnvironmentPipelineStageCloneRepositoriesNoStrategy() {

        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Prepare System': true
        ]
        
        jsr.step.abapEnvironmentPipelineStageCloneRepositories(script: nullScript)

        assertThat(stepsCalled, hasItems('abapEnvironmentPullGitRepo'))
        assertThat(stepsCalled, not(hasItems('abapEnvironmentCloneGitRepo', 'abapEnvironmentCheckoutBranch')))
    }
}
