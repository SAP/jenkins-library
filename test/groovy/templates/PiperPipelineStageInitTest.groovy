package templates

import com.sap.piper.StageNameProvider
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.hasItems
import static org.hamcrest.Matchers.hasKey
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.isEmptyOrNullString
import static org.hamcrest.Matchers.not
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue

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
    private Map  stepParams = [:]

    @Before
    void init() {
        StageNameProvider.instance.useTechnicalStageNames = false

        nullScript.env.STAGE_NAME = 'Init'

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
            stepParams['setupCommonPipelineEnvironment'] = m
        })

        helper.registerAllowedMethod('piperInitRunStageConfiguration', [Map.class], { m ->
            assertThat(m.stageConfigResource, not(isEmptyOrNullString()))
            stepsCalled.add('piperInitRunStageConfiguration')
        })

        helper.registerAllowedMethod('artifactSetVersion', [Map.class], { m ->
            stepsCalled.add('artifactSetVersion')
        })

        helper.registerAllowedMethod('artifactPrepareVersion', [Map.class], { m ->
            stepsCalled.add('artifactPrepareVersion')
        })

        helper.registerAllowedMethod('pipelineStashFilesBeforeBuild', [Map.class], { m ->
            stepsCalled.add('pipelineStashFilesBeforeBuild')
        })

        helper.registerAllowedMethod('slackSendNotification', [Map.class], {m ->
            stepsCalled.add('slackSendNotification')
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

        thrown.expectMessage('[piperPipelineStageInit] buildTool configuration \'npm\' does not fit to your project, please set buildTool as general setting in your .pipeline/config.yml correctly, see also https://sap.github.io/jenkins-library/configuration/')
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

        assertThat(stepsCalled, hasItems('checkout', 'setupCommonPipelineEnvironment', 'piperInitRunStageConfiguration', 'artifactPrepareVersion', 'pipelineStashFilesBeforeBuild'))
        assertThat(stepsCalled, not(hasItems('slackSendNotification')))

    }

    @Test
    void testInitNotOnProductiveBranch() {

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
    void testPullRequestStageStepActivation() {

        nullScript.commonPipelineEnvironment.configuration = [
            runStep: [:]
        ]
        def config = [
            pullRequestStageName: 'Pull-Request Voting',
            stepMappings        : [
                karma      : 'karmaExecuteTests',
                whitesource: 'whitesourceExecuteScan'
            ],
            labelPrefix         : 'pr_'
        ]

        def actions = ['karma', 'pr_whitesource']
        jsr.step.piperPipelineStageInit.setPullRequestStageStepActivation(nullScript, config, actions)

        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep."Pull-Request Voting".karmaExecuteTests, is(true))
        assertThat(nullScript.commonPipelineEnvironment.configuration.runStep."Pull-Request Voting".whitesourceExecuteScan, is(true))
    }

    @Test
    void testInitWithSlackNotification() {
        nullScript.commonPipelineEnvironment.configuration = [runStep: [Init: [slackSendNotification: true]]]

        jsr.step.piperPipelineStageInit(script: nullScript, juStabUtils: utils, buildTool: 'maven')

        assertThat(stepsCalled, hasItems(
            'checkout',
            'setupCommonPipelineEnvironment',
            'piperInitRunStageConfiguration',
            'artifactPrepareVersion',
            'slackSendNotification',
            'pipelineStashFilesBeforeBuild'
        ))
    }

    @Test
    void testInitInferBuildTool() {
        nullScript.commonPipelineEnvironment.configuration = [general: [inferBuildTool: true]]
        nullScript.commonPipelineEnvironment.buildTool = 'maven'

        jsr.step.piperPipelineStageInit(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItems(
            'checkout',
            'setupCommonPipelineEnvironment',
            'piperInitRunStageConfiguration',
            'artifactPrepareVersion',
            'pipelineStashFilesBeforeBuild'
        ))
    }

    @Test
    void testInitWithTechnicalStageNames() {
        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], { m, body ->
            assertThat(m.stageName, is('init'))
            return body()
        })

        jsr.step.piperPipelineStageInit(script: nullScript, juStabUtils: utils, useTechnicalStageNames: true, buildTool: 'maven')

        assertThat(stepsCalled, hasItems(
            'checkout',
            'setupCommonPipelineEnvironment',
            'piperInitRunStageConfiguration',
            'artifactPrepareVersion',
            'pipelineStashFilesBeforeBuild'
        ))
    }

    @Test
    void testInitForwardConfigParams() {
        jsr.step.piperPipelineStageInit(script: nullScript, juStabUtils: utils, configFile: 'my-config.yml',
            customDefaults: ['my-custom-defaults.yml'], customDefaultsFromFiles: ['my-custom-default-file.yml'],
            buildTool: 'maven')

        assertThat(stepsCalled, hasItems('setupCommonPipelineEnvironment'))
        assertThat(stepParams.setupCommonPipelineEnvironment?.configFile, is('my-config.yml'))
        assertThat(stepParams.setupCommonPipelineEnvironment?.customDefaults, is(['my-custom-defaults.yml']))
        assertThat(stepParams.setupCommonPipelineEnvironment?.customDefaultsFromFiles, is(['my-custom-default-file.yml']))
    }

    @Test
    void "Parameter skipCheckout skips the checkout call"() {
        jsr.step.piperPipelineStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven',
            stashSettings: 'com.sap.piper/pipeline/stashSettings.yml',
            skipCheckout: true,
            scmInfo: ["dummyScmKey":"dummyScmKey"]
        )

        assertThat(stepsCalled, hasItems('setupCommonPipelineEnvironment', 'piperInitRunStageConfiguration', 'artifactPrepareVersion', 'pipelineStashFilesBeforeBuild'))
        assertThat(stepsCalled, not(hasItem('checkout')))
    }

    @Test
    void "Try to skip checkout with parameter skipCheckout not boolean throws error"() {
        thrown.expectMessage('[piperPipelineStageInit] Parameter skipCheckout has to be of type boolean. Instead got \'java.lang.String\'')

        jsr.step.piperPipelineStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven',
            stashSettings: 'com.sap.piper/pipeline/stashSettings.yml',
            skipCheckout: "false"
        )
    }

    @Test
    void "Try to skip checkout without scmInfo parameter throws error"() {
        thrown.expectMessage('[piperPipelineStageInit] Need am scmInfo map retrieved from a checkout. ' +
            'If you want to skip the checkout the scm info needs to be provided by you with parameter scmInfo, ' +
            'for example as follows:\n' +
            '  def scmInfo = checkout scm\n' +
            '  piperPipelineStageInit script:this, skipCheckout: true, scmInfo: scmInfo')

        jsr.step.piperPipelineStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven',
            stashSettings: 'com.sap.piper/pipeline/stashSettings.yml',
            skipCheckout: true
        )
    }

}
