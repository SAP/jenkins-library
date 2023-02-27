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
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue

class PiperPipelineStageInitTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsReadYamlRule jryr = new JenkinsReadYamlRule(this)
    private JenkinsReadMavenPomRule jrmpr = new JenkinsReadMavenPomRule(this, null)
    private ExpectedException thrown = new ExpectedException()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jryr)
        .around(jrmpr)
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
                case 'mta.yaml':
                case 'path/mta.yaml':
                case 'pathFromStep/mta.yaml':
                case 'pathFromStage/mta.yaml':
                case 'pom.xml':
                case 'path/pom.xml':
                case 'pathFromStep/pom.xml':
                case 'pathFromStage/pom.xml':
                    return [new File(map.glob)].toArray()
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

        helper.registerAllowedMethod('transportRequestReqIDFromGit', [Map.class], {m ->
            stepsCalled.add('transportRequestReqIDFromGit')
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

        thrown.expectMessage('[piperPipelineStageInit] buildTool configuration \'npm\' does not fit to your project (buildDescriptorPattern: \'package.json\'), please set buildTool as general setting in your .pipeline/config.yml correctly, see also https://sap.github.io/jenkins-library/configuration/')
        jsr.step.piperPipelineStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'npm',
            stashSettings: 'com.sap.piper/pipeline/stashSettings.yml'
        )

    }
    
    @Test
    void testInitMtaBuildToolDoesNotThrowException() {

        jsr.step.piperPipelineStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'mta',
            stashSettings: 'com.sap.piper/pipeline/stashSettings.yml'
        )
        assertThat(stepsCalled, hasItems('checkout', 'setupCommonPipelineEnvironment', 'piperInitRunStageConfiguration', 'artifactPrepareVersion', 'pipelineStashFilesBeforeBuild'))
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
        assertThat(nullScript.commonPipelineEnvironment.configuration.stageStashes.Init.unstash, is([]))
    }

    @Test
    void testTransportRequestReqIDFromGitIfFalse() {
        jsr.step.piperPipelineStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven',
            stashSettings: 'com.sap.piper/pipeline/stashSettings.yml',
            transportRequestReqIDFromGit: false
        )
        assertThat(stepsCalled, not(hasItem('transportRequestReqIDFromGit')))
    }

    @Test
    void testTransportRequestReqIDFromGitIfTrue() {
        jsr.step.piperPipelineStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven',
            stashSettings: 'com.sap.piper/pipeline/stashSettings.yml',
            transportRequestReqIDFromGit: true
        )
        assertThat(stepsCalled, hasItem( 'transportRequestReqIDFromGit'))
    }

    @Test
    void testCustomStashSettings() {
        jryr.registerYaml('customStashSettings.yml',"Init: \n  unstash: source")

        jsr.step.piperPipelineStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven',
            customStashSettings: 'customStashSettings.yml',
            stashSettings: 'com.sap.piper/pipeline/stashSettings.yml'
        )

        assertThat(stepsCalled, hasItems('checkout', 'setupCommonPipelineEnvironment', 'piperInitRunStageConfiguration', 'artifactPrepareVersion', 'pipelineStashFilesBeforeBuild'))
        assertThat(stepsCalled, not(hasItems('slackSendNotification')))
        assertThat(nullScript.commonPipelineEnvironment.configuration.stageStashes.Init.unstash, is("source"))
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
    void testInferBuildToolDescMta() {
        assertEquals('mta.yaml', jsr.step.piperPipelineStageInit.inferBuildToolDesc(nullScript, "mta"))
    }

    @Test
    void testInferBuildToolDescMaven() {
        assertEquals('pom.xml', jsr.step.piperPipelineStageInit.inferBuildToolDesc(nullScript, "maven"))
    }

    @Test
    void testInferBuildToolDescNpm() {
        assertEquals('package.json', jsr.step.piperPipelineStageInit.inferBuildToolDesc(nullScript, "npm"))
    }

    @Test
    void testInferBuildToolDescMtaSource() {
        nullScript.commonPipelineEnvironment.configuration = [general: [buildTool: 'mta'], steps : [mtaBuild: [source: 'pathFromStep']]]

        helper.registerAllowedMethod('artifactPrepareVersion', [Map.class, Closure.class], { m, body ->
            assertThat(m.filePath, is('pathFromStep/mta.yaml'))
            return body()
        })

        jsr.step.piperPipelineStageInit(script: nullScript, juStabUtils: utils)

    }

    @Test
    void testInferBuildToolDescMtaSourceStage() {
        nullScript.commonPipelineEnvironment.configuration = [general: [buildTool: 'mta'], stages: [Build: [source: 'pathFromStage']], steps : [mtaBuild: [source: 'pathFromStep']]]

        helper.registerAllowedMethod('artifactPrepareVersion', [Map.class, Closure.class], { m, body ->
            assertThat(m.filePath, is('pathFromStage/mta.yaml'))
            return body()
        })

        jsr.step.piperPipelineStageInit(script: nullScript, juStabUtils: utils)
    }

    @Test
    void testInferBuildToolDescMavenSource() {
        nullScript.commonPipelineEnvironment.configuration = [general: [buildTool: 'maven'], steps : [mavenBuild: [pomPath: 'pathFromStep/pom.xml']]]

        helper.registerAllowedMethod('artifactPrepareVersion', [Map.class, Closure.class], { m, body ->
            assertThat(m.filePath, is('pathFromStep/pom.xml'))
            return body()
        })

        jsr.step.piperPipelineStageInit(script: nullScript, juStabUtils: utils)
    }

    @Test
    void testInferBuildToolDescMavenSourceStage() {
        nullScript.commonPipelineEnvironment.configuration = [general: [buildTool: 'maven'], stages: [Build: [pomPath: 'pathFromStage/pom.xml']], steps : [mavenBuild: [pomPath: 'pathFromStep/pom.xml']]]

        helper.registerAllowedMethod('artifactPrepareVersion', [Map.class, Closure.class], { m, body ->
            assertThat(m.filePath, is('pathFromStage/pom.xml'))
            return body()
        })

        jsr.step.piperPipelineStageInit(script: nullScript, juStabUtils: utils)
    }

    @Test
    void testInferBuildToolDescUnknown() {
        assertEquals(null, jsr.step.piperPipelineStageInit.inferBuildToolDesc(nullScript, "unknown"))
    }

    @Test
    void testInferBuildToolDescNull() {
        assertEquals(null, jsr.step.piperPipelineStageInit.inferBuildToolDesc(nullScript, null))
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
    void testInferProjectNameFromMta() {
        jryr.registerYaml('mta.yaml','ID: "fromMtaYaml"')
        assertEquals('fromMtaYaml', jsr.step.piperPipelineStageInit.inferProjectName(nullScript, "mta", "mta.yaml"))
    }

    @Test
    void testInferProjectNameFromMtaSource() {
        nullScript.commonPipelineEnvironment.configuration = [general: [buildTool: 'mta', inferProjectName: true], steps : [mtaBuild: [source: 'path']]]

        jryr.registerYaml('path/mta.yaml','ID: "fromPathMtaYaml"')
        jsr.step.piperPipelineStageInit(script: nullScript, juStabUtils: utils)
        assertEquals('fromPathMtaYaml', nullScript.commonPipelineEnvironment.projectName)
    }

    @Test
    void testInferProjectNameFromMavenPath() {
        jrmpr.registerPom('path/pom.xml',
            '<project>'
                + '<groupId>gidFromPathPom</groupId>'
                + '<artifactId>aidFromPathPom</artifactId>'
            + '</project>'
        )

        assertEquals('gidFromPathPom-aidFromPathPom', jsr.step.piperPipelineStageInit.inferProjectName(nullScript, "maven", "path/pom.xml"))
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
            stashContent: ['mystash'],
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
            stashContent: ['mystash'],
            skipCheckout: true
        )
    }

    @Test
    void "Try to skip checkout without stashContent parameter throws error"() {
        thrown.expectMessage('[piperPipelineStageInit] needs stashes if you skip checkout')

        jsr.step.piperPipelineStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven',
            stashSettings: 'com.sap.piper/pipeline/stashSettings.yml',
            skipCheckout: true,
            scmInfo: ["dummyScmKey":"dummyScmKey"]
        )
    }

    @Test
    void "Try to skip checkout with empty stashContent parameter throws error"() {
        thrown.expectMessage('[piperPipelineStageInit] needs stashes if you skip checkout')

        jsr.step.piperPipelineStageInit(
            script: nullScript,
            juStabUtils: utils,
            buildTool: 'maven',
            stashSettings: 'com.sap.piper/pipeline/stashSettings.yml',
            skipCheckout: true,
            stashContent: [],
            scmInfo: ["dummyScmKey":"dummyScmKey"]
        )
    }
}
