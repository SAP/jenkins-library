package templates

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class PiperPipelineStagePostTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(jsr)

    private List stepsCalled = []

    @Before
    void init()  {
        nullScript.env.STAGE_NAME = 'Release'

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Release'))
            return body()
        })
        helper.registerAllowedMethod('influxWriteData', [Map.class], {m -> stepsCalled.add('influxWriteData')})
        helper.registerAllowedMethod('slackSendNotification', [Map.class], {m -> stepsCalled.add('slackSendNotification')})
        helper.registerAllowedMethod('mailSendNotification', [Map.class], {m -> stepsCalled.add('mailSendNotification')})
        helper.registerAllowedMethod('piperPublishWarnings', [Map.class], {m -> stepsCalled.add('piperPublishWarnings')})
    }

    @Test
    void testPostDefault() {
        jsr.step.piperPipelineStagePost(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItems('influxWriteData','mailSendNotification','piperPublishWarnings'))
        assertThat(stepsCalled, not(hasItem('slackSendNotification')))
    }

    @Test
    void testPostNotOnProductiveBranch() {
        binding.variables.env.BRANCH_NAME = 'anyOtherBranch'

        jsr.step.piperPipelineStagePost(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItems('influxWriteData','mailSendNotification','piperPublishWarnings'))
        assertThat(stepsCalled, not(hasItems('slackSendNotification')))
    }

    @Test
    void testPostWithSlackNotification() {
        nullScript.commonPipelineEnvironment.configuration = [runStep: ['Post Actions': [slackSendNotification: true]]]

        jsr.step.piperPipelineStagePost(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItems('influxWriteData','mailSendNotification','slackSendNotification','piperPublishWarnings'))
    }
}
