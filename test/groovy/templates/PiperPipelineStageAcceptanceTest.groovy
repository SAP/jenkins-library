package templates

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.hasItems
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.junit.Assert.assertThat

class PiperPipelineStageAcceptanceTest extends BasePiperTest {
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
        binding.variables.env.STAGE_NAME = 'Acceptance'
        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Acceptance'))
            return body()
        })

        helper.registerAllowedMethod('healthExecuteCheck', [Map.class], {m ->
            stepsCalled.add('healthExecuteCheck')
            stepParameters.healthExecuteCheck = m
        })

        helper.registerAllowedMethod('cloudFoundryDeploy', [Map.class], {m ->
            stepsCalled.add('cloudFoundryDeploy')
            stepParameters.cloudFoundryDeploy = m
        })

        helper.registerAllowedMethod('neoDeploy', [Map.class], {m ->
            stepsCalled.add('neoDeploy')
            stepParameters.neoDeploy = m
        })

        helper.registerAllowedMethod('gaugeExecuteTests', [Map.class], {m ->
            stepsCalled.add('gaugeExecuteTests')
            stepParameters.gaugeExecuteTests = m
        })

        helper.registerAllowedMethod('newmanExecute', [Map.class], {m ->
            stepsCalled.add('newmanExecute')
            stepParameters.newmanExecute = m
        })

        helper.registerAllowedMethod('uiVeri5ExecuteTests', [Map.class], {m ->
            stepsCalled.add('uiVeri5ExecuteTests')
            stepParameters.uiVeri5ExecuteTests = m
        })

        helper.registerAllowedMethod('testsPublishResults', [Map.class], {m ->
            stepsCalled.add('testsPublishResults')
            stepParameters.testsPublishResults = m
        })
    }

    @Test
    void testAcceptanceStageDefault() {

        jsr.step.piperPipelineStageAcceptance(
            script: nullScript,
            juStabUtils: utils
        )
        assertThat(stepsCalled, not(hasItems('cloudFoundryDeploy', 'neoDeploy', 'healthExecuteCheck', 'newmanExecute', 'uiVeri5ExecuteTests', 'gaugeExecuteTests')))

    }

    @Test
    void testAcceptanceStageCF() {

        jsr.step.piperPipelineStageAcceptance(
            script: nullScript,
            juStabUtils: utils,
            cloudFoundryDeploy: true,
            healthExecuteCheck: true
        )

        assertThat(stepsCalled, hasItems('cloudFoundryDeploy', 'healthExecuteCheck'))
        assertThat(stepsCalled, not(hasItem('testsPublishResults')))
    }

    @Test
    void testAcceptanceStageNeo() {

        jsr.step.piperPipelineStageAcceptance(
            script: nullScript,
            juStabUtils: utils,
            neoDeploy: true
        )
        assertThat(stepsCalled, hasItem('neoDeploy'))
        assertThat(stepsCalled, not(hasItem('testsPublishResults')))
    }

    @Test
    void testAcceptanceStageGauge() {

        jsr.step.piperPipelineStageAcceptance(
            script: nullScript,
            juStabUtils: utils,
            gaugeExecuteTests: true
        )
        assertThat(stepsCalled, hasItems('gaugeExecuteTests', 'testsPublishResults'))
        assertThat(stepParameters.testsPublishResults.gauge.archive, is(true))
    }

    @Test
    void testAcceptanceStageNewman() {

        jsr.step.piperPipelineStageAcceptance(
            script: nullScript,
            juStabUtils: utils,
            newmanExecute: true
        )
        assertThat(stepsCalled, hasItems('newmanExecute', 'testsPublishResults'))
    }

    @Test
    void testAcceptanceStageUiVeri5() {

        jsr.step.piperPipelineStageAcceptance(
            script: nullScript,
            juStabUtils: utils,
            uiVeri5ExecuteTests: true
        )
        assertThat(stepsCalled, hasItems('uiVeri5ExecuteTests', 'testsPublishResults'))
    }
}
