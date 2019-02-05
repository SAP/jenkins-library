#!groovy
package templates

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.hasItems
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.isEmptyOrNullString
import static org.hamcrest.Matchers.not
import static org.junit.Assert.assertThat

class PiperPipelineStageInitTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private ExpectedException thrown = new ExpectedException()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(jlr)
        .around(jsr)

    private List stepsCalled = []

    @Before
    void init() {

        binding.variables.env.STAGE_NAME = 'Init'

        nullScript.commonPipelineEnvironment.configuration = [:]

        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            switch (map.glob) {
                case 'pom.xml':
                    return [new File('pom.xml')].toArray()
                default:
                    return [].toArray()
            }
        })

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], { m, body ->
            assertThat(m.stageName, is('Init'))
            return body()
        })

        helper.registerAllowedMethod('checkout', [Closure.class], { c ->
            stepsCalled.add('checkout')
            return [GIT_BRANCH: 'master', GIT_COMMIT: 'testGitCommitId', GIT_URL: 'https://github.com/testOrg/testRepo']
        })
        binding.setVariable('scm', {})

        helper.registerAllowedMethod('setupCommonPipelineEnvironment', [Map.class], { m ->
            stepsCalled.add('setupCommonPipelineEnvironment')
        })

        helper.registerAllowedMethod('piperInitRunStageConfiguration', [Map.class], { m ->
            assertThat(m.stageConfigResource, not(isEmptyOrNullString()))
            stepsCalled.add('piperInitRunStageConfiguration')
        })

        helper.registerAllowedMethod('artifactSetVersion', [Map.class], { m ->
            stepsCalled.add('artifactSetVersion')
        })

        helper.registerAllowedMethod('pipelineStashFilesBeforeBuild', [Map.class], { m ->
            stepsCalled.add('pipelineStashFilesBeforeBuild')
        })

    }

    @Test
    void testInitNoBuildTool() {

        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR buildTool')
        jsr.step.piperPipelineStageInit(
            script: nullScript,
            juStabUtils: utils,
            stashSettings: 'com.sap.piper/pipeline/stashSettings.yml'
        )

    }

    @Test
    void testInitBuildToolDoesNotMatchProject() {

        thrown.expectMessage('[piperPipelineStageInit] buildTool configuration \'npm\' does not fit to your project, please set buildTool as genereal setting in your .pipeline/config.yml correctly, see also https://github.wdf.sap.corp/pages/ContinuousDelivery/piper-doc/configuration/')
        jsr.step.piperPipelineStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'npm',
            stashSettings: 'com.sap.piper/pipeline/stashSettings.yml'
        )

    }

    @Test
    void testInitDefault() {

        jsr.step.piperPipelineStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven',
            stashSettings: 'com.sap.piper/pipeline/stashSettings.yml'
        )

        assertThat(stepsCalled, hasItems('checkout', 'setupCommonPipelineEnvironment', 'piperInitRunStageConfiguration', 'artifactSetVersion', 'pipelineStashFilesBeforeBuild'))

    }

    @Test
    void testInitOverwriteDefault() {

        binding.variables.env.BRANCH_NAME = 'testBranch'

        jsr.step.piperPipelineStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven',
            stashSettings: 'com.sap.piper/pipeline/stashSettings.yml'
        )

        assertThat(stepsCalled, hasItems('checkout', 'setupCommonPipelineEnvironment', 'piperInitRunStageConfiguration', 'pipelineStashFilesBeforeBuild'))
        assertThat(stepsCalled, not(hasItems('artifactSetVersion')))

    }

    @Test
    void testSetScmInfoOnCommonPipelineEnvironment() {
        //currently supported formats
        def scmInfoTestList = [
            [GIT_URL: 'https://github.com/testOrg/testRepo.git', expectedSsh: 'git@github.com:testOrg/testRepo.git', expectedHttp: 'https://github.com/testOrg/testRepo.git'],
            [GIT_URL: 'https://github.com:7777/testOrg/testRepo.git', expectedSsh: 'git@github.com:testOrg/testRepo.git', expectedHttp: 'https://github.com:7777/testOrg/testRepo.git'],
            [GIT_URL: 'git@github.com:testOrg/testRepo.git', expectedSsh: 'git@github.com:testOrg/testRepo.git', expectedHttp: 'https://github.com/testOrg/testRepo.git'],
            [GIT_URL: 'ssh://git@github.com/testOrg/testRepo.git', expectedSsh: 'ssh://git@github.com/testOrg/testRepo.git', expectedHttp: 'https://github.com/testOrg/testRepo.git'],
            [GIT_URL: 'ssh://git@github.com:7777/testOrg/testRepo.git', expectedSsh: 'ssh://git@github.com:7777/testOrg/testRepo.git', expectedHttp: 'https://github.com/testOrg/testRepo.git'],
        ]

        scmInfoTestList.each {scmInfoTest ->
            jsr.step.piperPipelineStageInit.setScmInfoOnCommonPipelineEnvironment(nullScript, scmInfoTest)
            assertThat(nullScript.commonPipelineEnvironment.getGitSshUrl(), is(scmInfoTest.expectedSsh))
            assertThat(nullScript.commonPipelineEnvironment.getGitHttpsUrl(), is(scmInfoTest.expectedHttp))
        }
    }
}
