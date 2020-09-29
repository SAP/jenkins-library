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
import static org.junit.Assert.assertTrue

class abapEnvironmentPipelineStageInitTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this).registerYaml('mta.yaml', defaultsYaml() )
    private List stepsCalled = []
    private List activeStages = []

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(readYamlRule)
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
    void testStageConfiguration() {

        jsr.step.abapEnvironmentPipelineStageInit(script: nullScript)
        assertThat(stepsCalled, hasItems('setupCommonPipelineEnvironment', 'checkout'))

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
