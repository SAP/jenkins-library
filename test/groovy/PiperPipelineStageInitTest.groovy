#!groovy
package stages

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class PiperPipelineStageInitTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(jlr)
        .around(jsr)

    private List stepsCalled = []

    @Before
    void init()  {
        binding.variables.env.STAGE_NAME = 'Init'
        binding.setVariable('scm', {})

        helper.registerAllowedMethod('deleteDir', [], null)
        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            switch (map.glob) {
                case 'pom.xml':
                    return [new File('pom.xml')].toArray()
                default:
                    return [].toArray()
            }
        })
        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Init'))
            return body()
        })
        helper.registerAllowedMethod('checkout', [Closure.class], {c ->
            stepsCalled.add('checkout')
            return [
                GIT_COMMIT: 'abcdef12345',
                GIT_URL: 'some.url'
            ]
        })
        helper.registerAllowedMethod('setupCommonPipelineEnvironment', [Map.class], {m -> stepsCalled.add('setupCommonPipelineEnvironment')})
        helper.registerAllowedMethod('piperInitRunStageConfiguration', [Map.class], {m -> stepsCalled.add('piperInitRunStageConfiguration')})
        helper.registerAllowedMethod('slackSendNotification', [Map.class], {m -> stepsCalled.add('slackSendNotification')})
        helper.registerAllowedMethod('artifactSetVersion', [Map.class], {m -> stepsCalled.add('artifactSetVersion')})
        helper.registerAllowedMethod('pipelineStashFilesBeforeBuild', [Map.class], {m -> stepsCalled.add('pipelineStashFilesBeforeBuild')})
    }

    @Test
    void testInitNoBuildTool() {
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR buildTool')
        jsr.step.piperPipelineStageInit(script: nullScript, juStabUtils: utils)
    }

    @Test
    void testInitBuildToolDoesNotMatchProject() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage(containsString("buildTool configuration 'npm' does not fit to your project"))
        jsr.step.piperPipelineStageInit(script: nullScript, juStabUtils: utils, buildTool: 'npm')
    }

    @Test
    void testInitDefault() {
        jsr.step.piperPipelineStageInit(script: nullScript, juStabUtils: utils, buildTool: 'maven')

        assertThat(stepsCalled, hasItems(
            'checkout',
            'setupCommonPipelineEnvironment',
            'piperInitRunStageConfiguration',
            'artifactSetVersion',
            'pipelineStashFilesBeforeBuild'
        ))
        assertThat(stepsCalled, not(hasItems('slackSendNotification')))
    }

    @Test
    void testInitNotOnProductiveBranch() {
        binding.variables.env.BRANCH_NAME = 'anyOtherBranch'

        jsr.step.piperPipelineStageInit(script: nullScript, juStabUtils: utils, buildTool: 'maven')

        assertThat(stepsCalled, hasItems(
            'checkout',
            'setupCommonPipelineEnvironment',
            'piperInitRunStageConfiguration',
            'pipelineStashFilesBeforeBuild'
        ))
        assertThat(stepsCalled, not(hasItems('artifactSetVersion')))
    }

    @Test
    void testInitWithSlackNotification() {
        nullScript.commonPipelineEnvironment.configuration = [runStep: [Init: [slackSendNotification: true]]]

        jsr.step.piperPipelineStageInit(script: nullScript, juStabUtils: utils, buildTool: 'maven')

        assertThat(stepsCalled, hasItems(
            'checkout',
            'setupCommonPipelineEnvironment',
            'piperInitRunStageConfiguration',
            'artifactSetVersion',
            'slackSendNotification',
            'pipelineStashFilesBeforeBuild'
        ))
    }
}
