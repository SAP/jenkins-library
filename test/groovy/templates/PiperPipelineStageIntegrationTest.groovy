package templates

import com.sap.piper.Utils

import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class PiperPipelineStageIntegrationTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    private List stepsCalled = []
    private Map stepParameters = [:]

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jlr)
        .around(jsr)

    @Before
    void init()  {
        binding.variables.env.STAGE_NAME = 'Integration'
        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            return body()
        })

        helper.registerAllowedMethod('npmExecuteScripts', [Map.class], {m ->
            stepsCalled.add('npmExecuteScripts')
            stepParameters.npmExecuteScripts = m
        })

        helper.registerAllowedMethod('testsPublishResults', [Map.class], {m ->
            stepsCalled.add('testsPublishResults')
            stepParameters.testsPublishResults = m
        })

        helper.registerAllowedMethod('withEnv', [List.class, Closure.class], {env, body ->
            body()
        })

        Utils.metaClass.echo = { m -> }
    }

    @After
    void tearDown() {
        Utils.metaClass = null
    }

    @Test
    void testStageDefault() {
        jsr.step.piperPipelineStageIntegration(
            script: nullScript,
            juStabUtils: utils,
        )
        assertThat(stepsCalled, not(anyOf(hasItem('npmExecuteScripts'), hasItem('testsPublishResults'))))
    }

    @Test
    void testAcceptanceStageNpmExecuteScripts() {
        jsr.step.piperPipelineStageIntegration(
            script: nullScript,
            juStabUtils: utils,
            npmExecuteScripts: true
        )
        assertThat(stepsCalled, hasItems('npmExecuteScripts', 'testsPublishResults'))
    }
}
