package templates

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.PipelineWhenException
import org.junit.rules.ExpectedException
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue

class abapEnvironmentPipelineStageInitTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this).registerYaml('mta.yaml', defaultsYaml() )
    private List stepsCalled = []
    private List activeStages = []
    private ExpectedException thrown = new ExpectedException()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(readYamlRule)
        .around(thrown)
        .around(jsr)

    @Before
    void init()  {
        binding.variables.env.STAGE_NAME = 'Init'

        helper.registerAllowedMethod('deleteDir', [], null)

        helper.registerAllowedMethod('setupCommonPipelineEnvironment', [Map.class], { m ->
            stepsCalled.add('setupCommonPipelineEnvironment')
        })

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Init'))
            return body()
        })

        helper.registerAllowedMethod('checkout', [Closure.class], { c ->
            stepsCalled.add('checkout')
            return [GIT_BRANCH: 'master', GIT_COMMIT: 'testGitCommitId', GIT_URL: 'https://github.com/testOrg/testRepo']
        })
        binding.setVariable('scm', {})

        helper.registerAllowedMethod('activateStage', [Map.class, String.class], {p, m ->
            stepsCalled('activateStage')
            activeStages.add(m)
        })

        nullScript.prepareDefaultValues(script: nullScript)
    }

    @Test
    void testStageConfigurationToggleFalse() {
        jsr.step.abapEnvironmentPipelineStageInit(script: nullScript, skipCheckout: false)
        assertThat(stepsCalled, hasItems('setupCommonPipelineEnvironment', 'checkout'))
    }

    @Test
    void testSkipCheckoutToggleTrue() {
        jsr.step.abapEnvironmentPipelineStageInit(script: nullScript,  skipCheckout: true)
        assertThat(stepsCalled, not(hasItems('checkout')))
        assertThat(stepsCalled, hasItems('setupCommonPipelineEnvironment'))
    }

    @Test
    void testSkipCheckoutToggleNull() {
        jsr.step.abapEnvironmentPipelineStageInit(script: nullScript,  skipCheckout: null)
        assertThat(stepsCalled, hasItems('setupCommonPipelineEnvironment', 'checkout'))
    }

    @Test
    void testSkipCheckoutToggleString() {
        thrown.expectMessage('[abapEnvironmentPipelineStageInit] Parameter skipCheckout has to be of type boolean. Instead got \'java.lang.String\'')
        jsr.step.abapEnvironmentPipelineStageInit(script: nullScript,  skipCheckout: 'string')
    }

    private defaultsYaml() {
        return  '''
                stages:
                    Init: {}
                    Prepare System: {}
                    Clone Repositories: {}
                    ATC: {}
                '''
    }
}
