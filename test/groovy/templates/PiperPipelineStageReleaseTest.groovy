package templates

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.hasItems
import static org.hamcrest.Matchers.hasItems
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.hamcrest.Matchers.not
import static org.junit.Assert.assertThat

class PiperPipelineStageReleaseTest extends BasePiperTest {
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
        binding.variables.env.STAGE_NAME = 'Release'
        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Release'))
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

        helper.registerAllowedMethod('githubPublishRelease', [Map.class], {m ->
            stepsCalled.add('githubPublishRelease')
            stepParameters.githubPublishRelease = m
        })
    }

    @Test
    void testReleaseStageDefault() {

        jsr.step.piperPipelineStageRelease(
            script: nullScript,
            juStabUtils: utils
        )
        assertThat(stepsCalled, not(hasItems('cloudFoundryDeploy', 'neoDeploy', 'healthExecuteCheck', 'githubPublishRelease')))
    }

    @Test
    void testReleaseStageCF() {

        jsr.step.piperPipelineStageRelease(
            script: nullScript,
            juStabUtils: utils,
            cloudFoundryDeploy: true,
            healthExecuteCheck: true
        )

        assertThat(stepsCalled, hasItems('cloudFoundryDeploy', 'healthExecuteCheck'))
    }

    @Test
    void testReleaseStageNeo() {

        jsr.step.piperPipelineStageRelease(
            script: nullScript,
            juStabUtils: utils,
            neoDeploy: true
        )

        assertThat(stepsCalled, hasItem('neoDeploy'))
    }

    @Test
    void testReleaseStageGitHub() {

        jsr.step.piperPipelineStageRelease(
            script: nullScript,
            juStabUtils: utils,
            githubPublishRelease: true
        )

        assertThat(stepsCalled, hasItem('githubPublishRelease'))
    }
}
