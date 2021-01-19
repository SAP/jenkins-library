package templates

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.anyOf
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.junit.Assert.assertThat

class PiperPipelineStagePromoteTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jlr)
        .around(jsr)

    private List stepsCalled = []
    private Map stepParameters = [:]

    @Before
    void init()  {
        nullScript.env.STAGE_NAME = 'Promote'
        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Promote'))

            return body()
        })

        helper.registerAllowedMethod('containerPushToRegistry', [Map.class], {m ->
            stepsCalled.add('containerPushToRegistry')
            stepParameters.containerPushToRegistry = m
        })

        helper.registerAllowedMethod('nexusUpload', [Map.class], {m ->
            stepsCalled.add('nexusUpload')
            stepParameters.nexusUpload = m
        })
    }

    @Test
    void testStagePromoteDefault() {

        jsr.step.piperPipelineStagePromote(
            script: nullScript,
            juStabUtils: utils,
        )
        assertThat(stepsCalled, not(anyOf(hasItem('containerPushToRegistry'), hasItem('nexusUpload'))))

    }

    @Test
    void testStagePromotePushToRegistry() {

        jsr.step.piperPipelineStagePromote(
            script: nullScript,
            juStabUtils: utils,
            containerPushToRegistry: true
        )

        assertThat(stepsCalled, hasItem('containerPushToRegistry'))
    }

    @Test
    void testStagePromoteNexusUpload() {

        jsr.step.piperPipelineStagePromote(
            script: nullScript,
            juStabUtils: utils,
            nexusUpload: true
        )

        assertThat(stepsCalled, hasItem('nexusUpload'))
    }
}
