package templates

import org.junit.Before
import org.junit.After
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.PipelineWhenException
import org.junit.rules.ExpectedException
import util.Rules
import util.JenkinsShellCallRule
import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue

class abapEnvironmentPipelineStageInitTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this).registerYaml('mta.yaml', defaultsYaml() )
    private List stepsCalled = []
    private List activeStages = []
    private ExpectedException thrown = new ExpectedException()
    private JenkinsShellCallRule shellCallRule = new JenkinsShellCallRule(this)
    private PiperGoUtils piperGoUtils = new PiperGoUtils(nullScript, utils) { void unstashPiperBin() { }}

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(readYamlRule)
        .around(thrown)
        .around(jsr)
        .around(shellCallRule)

    @After
    public void tearDown() {
        Utils.metaClass = null
    }

    @Before
    void init()  {
        Utils.metaClass.unstash = { def n -> ["dummy"] }
        binding.variables.env.STAGE_NAME = 'Init'

        helper.registerAllowedMethod('deleteDir', [], null)
        helper.registerAllowedMethod("writeFile", [Map.class], null)
        helper.registerAllowedMethod("readJSON", [Map.class],null)

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
        shellCallRule.setReturnValue('[ -x ./piper ]', 1)
        shellCallRule.setReturnValue('./piper checkIfStepActive --stageConfig .pipeline/stage_conditions.yaml --useV1 --stageOutputFile .pipeline/stage_out.json --stepOutputFile .pipeline/step_out.json --stage _ --step _', 0)
        nullScript.prepareDefaultValues(script: nullScript)
    }

    @Test
    void testStageConfigurationToggleFalse() {
        jsr.step.abapEnvironmentPipelineStageInit(script: nullScript, skipCheckout: false, piperGoUtils: piperGoUtils)
        assertThat(stepsCalled, hasItems('setupCommonPipelineEnvironment', 'checkout'))
        assertThat(shellCallRule.shell, hasItem('./piper checkIfStepActive --stageConfig .pipeline/stage_conditions.yaml --useV1 --stageOutputFile .pipeline/stage_out.json --stepOutputFile .pipeline/step_out.json --stage _ --step _'))

    }

    @Test
    void testSkipCheckoutToggleTrue() {
        jsr.step.abapEnvironmentPipelineStageInit(
            script: nullScript,
            skipCheckout: true,
            juStabUtils: utils,
            piperGoUtils: piperGoUtils,
            stashContent: ['mystash']
        )
        assertThat(stepsCalled, not(hasItems('checkout')))
        assertThat(stepsCalled, hasItems('setupCommonPipelineEnvironment'))
        assertThat(shellCallRule.shell, hasItem('./piper checkIfStepActive --stageConfig .pipeline/stage_conditions.yaml --useV1 --stageOutputFile .pipeline/stage_out.json --stepOutputFile .pipeline/step_out.json --stage _ --step _'))

    }

    @Test
    void testSkipCheckoutToggleNull() {
        jsr.step.abapEnvironmentPipelineStageInit(script: nullScript,  skipCheckout: null, piperGoUtils: piperGoUtils)
        assertThat(stepsCalled, hasItems('setupCommonPipelineEnvironment', 'checkout'))
        assertThat(shellCallRule.shell, hasItem('./piper checkIfStepActive --stageConfig .pipeline/stage_conditions.yaml --useV1 --stageOutputFile .pipeline/stage_out.json --stepOutputFile .pipeline/step_out.json --stage _ --step _'))
    }

    @Test
    void testSkipCheckoutToggleString() {
        thrown.expectMessage('[abapEnvironmentPipelineStageInit] Parameter skipCheckout has to be of type boolean. Instead got \'java.lang.String\'')
        jsr.step.abapEnvironmentPipelineStageInit(script: nullScript,  skipCheckout: 'string')
    }

    @Test
    void "Try to skip checkout with parameter skipCheckout not boolean throws error"() {
        thrown.expectMessage('[abapEnvironmentPipelineStageInit] Parameter skipCheckout has to be of type boolean. Instead got \'java.lang.String\'')

        jsr.step.abapEnvironmentPipelineStageInit(
            script: nullScript,
            juStabUtils: utils,
            skipCheckout: "false",
            piperGoUtils: piperGoUtils
        )
    }

    @Test
    void "Try to skip checkout without stashContent parameter throws error"() {
        thrown.expectMessage('[abapEnvironmentPipelineStageInit] needs stashes if you skip checkout')

        jsr.step.abapEnvironmentPipelineStageInit(
            script: nullScript,
            juStabUtils: utils,
            skipCheckout: true,
            piperGoUtils: piperGoUtils
        )
    }

    @Test
    void "Try to skip checkout with empty stashContent parameter throws error"() {
        thrown.expectMessage('[abapEnvironmentPipelineStageInit] needs stashes if you skip checkout')

        jsr.step.abapEnvironmentPipelineStageInit(
            script: nullScript,
            juStabUtils: utils,
            skipCheckout: true,
            stashContent: [],
            piperGoUtils: piperGoUtils
        )
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
