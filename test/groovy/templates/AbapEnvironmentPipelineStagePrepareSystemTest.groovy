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

class AbapEnvironmentPipelineStagePrepareSystemTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jsr)

    private stepsCalled = []

    @Before
    void init()  {
        helper.registerAllowedMethod('input', [Map], {m -> return null})
        helper.registerAllowedMethod('cloudFoundryCreateService', [Map.class], {m -> stepsCalled.add('cloudFoundryCreateService')})
        helper.registerAllowedMethod('cloudFoundryCreateServiceKey', [Map.class], {m -> stepsCalled.add('cloudFoundryCreateServiceKey')})
    }

    @Test
    void testAbapEnvironmentPipelineStagePrepareSystemExecuted() {

        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Prepare System': true
        ]
        jsr.step.abapEnvironmentPipelineStagePrepareSystem(script: nullScript)

        assertThat(stepsCalled, hasItems('cloudFoundryCreateService', 'cloudFoundryCreateServiceKey'))
    }
}
