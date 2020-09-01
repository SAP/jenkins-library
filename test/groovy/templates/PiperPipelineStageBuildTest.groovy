package templates

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

import static org.hamcrest.Matchers.hasItems
import static org.hamcrest.Matchers.is
import static org.junit.Assert.assertThat

class PiperPipelineStageBuildTest extends BasePiperTest {
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

        nullScript.env.STAGE_NAME = 'Build'

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Build'))
            return body()
        })

        helper.registerAllowedMethod('buildExecute', [Map.class], {m ->
            stepsCalled.add('buildExecute')
        })

        helper.registerAllowedMethod('pipelineStashFilesAfterBuild', [Map.class], {m ->
            stepsCalled.add('pipelineStashFilesAfterBuild')
        })

        helper.registerAllowedMethod('checksPublishResults', [Map.class], {m ->
            stepsCalled.add('checksPublishResults')
        })

        helper.registerAllowedMethod('testsPublishResults', [Map.class], {m ->
            stepsCalled.add('testsPublishResults')
            stepParameters.testsPublishResults = m
        })
    }

    @Test
    void testBuildDefault() {

        jsr.step.piperPipelineStageBuild(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItems('buildExecute', 'checksPublishResults', 'pipelineStashFilesAfterBuild', 'testsPublishResults'))
        assertThat(stepParameters.testsPublishResults.junit.updateResults, is(true))
    }
}
