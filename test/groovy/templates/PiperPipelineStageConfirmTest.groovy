package templates

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.is
import static org.junit.Assert.assertThat

class PiperPipelineStageConfirmTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    private Map timeoutSettings
    private Map inputSettings
    private milestoneCalled = false

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jlr)
        .around(jsr)

    @Before
    void init()  {
        binding.variables.env.STAGE_NAME = 'Confirm'

        helper.registerAllowedMethod('timeout', [Map.class, Closure.class], {m, body ->
            timeoutSettings = m
            return body()
        })

        helper.registerAllowedMethod('input', [Map.class], {m ->
            inputSettings = m
            return [reason: 'this is my test reason for failing step 1 and step 3', acknowledgement: true]
        })

        helper.registerAllowedMethod('milestone', [],{
            milestoneCalled = true
        })
    }

    @Test
    void testStageDefault() {

        jsr.step.piperPipelineStageConfirm(
            script: nullScript
        )
        assertThat(timeoutSettings.unit, is('HOURS'))
        assertThat(timeoutSettings.time, is(720))
        assertThat(inputSettings.message, is('Shall we proceed to Promote & Release?'))
    }

    @Test
    void testStageBuildUnstable() {

        binding.setVariable('currentBuild', [result: 'UNSTABLE'])
        nullScript.commonPipelineEnvironment.setValue('unstableSteps', ['step1', 'step3'])

        helper.registerAllowedMethod('text', [Map.class], {m ->
            assertThat(m.defaultValue, is(''))
            assertThat(m.description, containsString('Please provide a reason for overruling the failed steps step1, step3, with 10 characters or more:'))
            assertThat(m.name, is('reason'))
        })

        helper.registerAllowedMethod('booleanParam', [Map.class], {m ->
            assertThat(m.description, is('I acknowledge that for traceability purposes the approval reason is stored together with my user name / user id:'))
            assertThat(m.name, is('acknowledgement'))
        })

        jsr.step.piperPipelineStageConfirm(
            script: nullScript
        )
        assertThat(inputSettings.message, is('Approve continuation of pipeline, although some steps failed.'))

        assertThat(jlr.log, containsString('step1'))
        assertThat(jlr.log, containsString('step3'))
        assertThat(jlr.log, containsString('this is my test reason'))
        assertThat(jlr.log, containsString('Acknowledgement\n---------------\n☑ I acknowledge that for traceability purposes the approval reason is stored together with my user name / user id'))
    }

    @Test
    void callsMilestone(){
        jsr.step.piperPipelineStageConfirm(
            script: nullScript
        )

        assertThat(milestoneCalled, is(true))
    }
}
