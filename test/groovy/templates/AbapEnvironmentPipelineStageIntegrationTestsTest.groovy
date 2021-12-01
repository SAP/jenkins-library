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
import static org.hamcrest.Matchers.equalTo;
import static org.hamcrest.Matchers.is;
import static org.junit.Assert.fail

class abapEnvironmentPipelineStageIntegrationTestsTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jsr)

    private stepsCalled = []

    @Before
    void init()  {
        binding.variables.env.STAGE_NAME = 'Integration Tests'

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Integration Tests'))
            return body()
        })
        helper.registerAllowedMethod('input', [Map], {m ->
            stepsCalled.add('input')
            return null
        })
        helper.registerAllowedMethod('abapEnvironmentCreateSystem', [Map.class], {m -> stepsCalled.add('abapEnvironmentCreateSystem')})
        helper.registerAllowedMethod('cloudFoundryDeleteService', [Map.class], {m -> stepsCalled.add('cloudFoundryDeleteService')})
    }

    @Test
    void testCloudFoundryDeleteServiceExecutedConfirm() {

        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Integration Tests': true
        ]
        jsr.step.abapEnvironmentPipelineStageIntegrationTests(script: nullScript, confirmDeletion: true)

        assertThat(stepsCalled, hasItems('input'))
        assertThat(stepsCalled, hasItems('abapEnvironmentCreateSystem'))
        assertThat(stepsCalled, hasItems('cloudFoundryDeleteService'))
    }

    @Test
    void testCloudFoundryDeleteServiceExecutedNoConfirm() {

        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Integration Tests': true
        ]
        jsr.step.abapEnvironmentPipelineStageIntegrationTests(script: nullScript, confirmDeletion: false)


        assertThat(stepsCalled, not(hasItem('input')))
        assertThat(stepsCalled, hasItems('abapEnvironmentCreateSystem'))
        assertThat(stepsCalled, hasItems('cloudFoundryDeleteService'))
    }

    @Test
    void testCreateSystemFails() {

        helper.registerAllowedMethod('abapEnvironmentCreateSystem', [Map.class], {m -> stepsCalled.add('abapEnvironmentCreateSystem'); error("Failed")})

        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Integration Tests': true
        ]

        try {
            jsr.step.abapEnvironmentPipelineStageIntegrationTests(script: nullScript, confirmDeletion: false)
            fail("Expected exception")
        } catch (Exception e) {
            // failure expected
        }

        assertThat(stepsCalled, not(hasItem('input')))
        assertThat(stepsCalled, hasItems('abapEnvironmentCreateSystem'))
        assertThat(stepsCalled, hasItems('cloudFoundryDeleteService'))
    }

}
