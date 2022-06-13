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

class AbapEnvironmentPipelineStageATCTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jsr)

    private stepsCalled = []

    @Before
    void init()  {
        binding.variables.env.STAGE_NAME = 'ATC'

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('ATC'))
            return body()
        })
        helper.registerAllowedMethod('abapEnvironmentRunATCCheck', [Map.class], {m -> stepsCalled.add('abapEnvironmentRunATCCheck')})
        helper.registerAllowedMethod('abapEnvironmentPushATCSystemConfig', [Map.class], {m -> stepsCalled.add('abapEnvironmentPushATCSystemConfig')})
        helper.registerAllowedMethod('cloudFoundryCreateServiceKey', [Map.class], {m -> stepsCalled.add('cloudFoundryCreateServiceKey')})
    }

    @Test
    void testAbapEnvironmentRunTests() {

        nullScript.commonPipelineEnvironment.configuration.runStage = []
        jsr.step.abapEnvironmentPipelineStageATC(script: nullScript)

        assertThat(stepsCalled, hasItems('abapEnvironmentRunATCCheck','cloudFoundryCreateServiceKey'))
        assertThat(stepsCalled, not(hasItems('abapEnvironmentPushATCSystemConfig')))
    }

    @Test
    void testAbapEnvironmentRunTestsWithATCSystemConfig() {

        nullScript.commonPipelineEnvironment.configuration.runStage = []
        jsr.step.abapEnvironmentPipelineStageATC(script: nullScript, atcSystemConfigFilePath: 'atcSystemConfig.json' )

        assertThat(stepsCalled, hasItems('abapEnvironmentRunATCCheck','abapEnvironmentPushATCSystemConfig','cloudFoundryCreateServiceKey'))
    }

    @Test
    void testAbapEnvironmentRunTestsWithHost() {

        nullScript.commonPipelineEnvironment.configuration.runStage = []
        jsr.step.abapEnvironmentPipelineStageATC(script: nullScript,  host: 'abc.com')

        assertThat(stepsCalled, hasItems('abapEnvironmentRunATCCheck'))
        assertThat(stepsCalled, not(hasItems('abapEnvironmentPushATCSystemConfig','cloudFoundryCreateServiceKey')))
    }

}
