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

import static org.hamcrest.Matchers.anyOf
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.hasItems
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.hamcrest.Matchers.nullValue
import static org.junit.Assert.assertThat

class PiperPipelineStagePRVotingTest extends BasePiperTest {
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

        nullScript.env.STAGE_NAME = 'Pull-Request Voting'
        binding.variables.env = [
            STAGE_NAME: 'Pull-Request Voting',
            BRANCH_NAME: 'PR-1'
        ]

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Pull-Request Voting'))
            return body()
        })

        helper.registerAllowedMethod('buildExecute', [Map.class], {m ->
            stepsCalled.add('buildExecute')
            stepParameters.buildExecute = m
        })

        helper.registerAllowedMethod('checksPublishResults', [Map.class], {m ->
            stepsCalled.add('checksPublishResults')
        })

        helper.registerAllowedMethod('testsPublishResults', [Map.class], {m ->
            stepsCalled.add('testsPublishResults')
        })

        helper.registerAllowedMethod('karmaExecuteTests', [Map.class], {m ->
            stepsCalled.add('karmaExecuteTests')
        })

        helper.registerAllowedMethod('mavenExecuteIntegration', [Map.class], {m ->
            stepsCalled.add('mavenExecuteIntegration')
        })

        helper.registerAllowedMethod('mavenExecuteStaticCodeChecks', [Map.class], {m ->
            stepsCalled.add('mavenExecuteStaticCodeChecks')
        })

        helper.registerAllowedMethod('npmExecuteLint', [Map.class], {m ->
            stepsCalled.add('npmExecuteLint')
        })

        helper.registerAllowedMethod('npmExecuteScripts', [Map.class], {m ->
            stepsCalled.add('npmExecuteScripts')
        })

        helper.registerAllowedMethod('whitesourceExecuteScan', [Map.class], {m ->
            stepsCalled.add('whitesourceExecuteScan')
            stepParameters.whitesourceExecuteScan = m
            m.script.commonPipelineEnvironment.setValue('whitesourceProjectNames', ['ws project - PR1'])

        })

        helper.registerAllowedMethod('writeTemporaryCredentials', [Map.class, Closure.class], {m, body ->
            stepsCalled.add('writeTemporaryCredentials')
            body()
        })
    }

    @Test
    void testPRVotingDefault() {

        nullScript.commonPipelineEnvironment.configuration = [general: [buildTool: 'maven']]
        jsr.step.piperPipelineStagePRVoting(script: nullScript, juStabUtils: utils)

        assertThat(stepsCalled, hasItems('buildExecute', 'checksPublishResults', 'testsPublishResults'))
        assertThat(stepsCalled, not(anyOf(hasItem('karmaExecuteTests'), hasItem('mavenExecuteIntegration'), hasItem('mavenExecuteStaticCodeChecks'), hasItem('npmExecuteLint'), hasItem('npmExecuteScripts'), hasItem('whitesourceExecuteScan'))))
        assertThat(stepParameters.buildExecute.buildTool, is('maven'))
        assertThat(stepParameters.buildExecute.dockerRegistryUrl, nullValue())
    }

    @Test
    void testPRVotingWithCustomSteps() {

        nullScript.commonPipelineEnvironment.configuration = [
            general: [buildTool: 'maven'],
            runStep: ['Pull-Request Voting': [karmaExecuteTests: true, whitesourceExecuteScan: true]]
        ]

        jsr.step.piperPipelineStagePRVoting(
            script: nullScript,
            juStabUtils: utils,
        )

        assertThat(stepsCalled, hasItems( 'karmaExecuteTests', 'whitesourceExecuteScan'))
        assertThat(stepParameters.whitesourceExecuteScan.productVersion, is('PR-1'))
    }

    @Test
    void testPRVotingDocker() {

        nullScript.commonPipelineEnvironment.configuration = [
            general: [buildTool: 'docker'],
            runStep: ['Pull-Request Voting': [karmaExecuteTests: true, whitesourceExecuteScan: true]]
        ]

        jsr.step.piperPipelineStagePRVoting(
            script: nullScript,
            juStabUtils: utils,
        )

        assertThat(stepParameters.buildExecute.dockerRegistryUrl, is(''))
    }

    @Test
    void testPRVotingWithMavenExecuteIntegration() {

        nullScript.commonPipelineEnvironment.configuration = [
            general: [buildTool: 'maven'],
            runStep: ['Pull-Request Voting': [mavenExecuteIntegration: true]]
        ]

        jsr.step.piperPipelineStagePRVoting(
            script: nullScript,
            juStabUtils: utils,
        )

        assertThat(stepsCalled, hasItems('mavenExecuteIntegration', 'testsPublishResults', 'writeTemporaryCredentials'))
    }

    @Test
    void testPRVotingWithLinting() {

        nullScript.commonPipelineEnvironment.configuration = [
            general: [buildTool: 'maven'],
            stages: ['Pull-Request Voting': [npmExecuteLint: true, mavenExecuteStaticCodeChecks: true]]
        ]

        jsr.step.piperPipelineStagePRVoting(
            script: nullScript,
            juStabUtils: utils,
        )

        assertThat(stepsCalled, hasItem('npmExecuteLint'))
        assertThat(stepsCalled, hasItem('mavenExecuteStaticCodeChecks'))

    }

    @Test
    void testPRVotingWithNpmExecuteScripts() {

        nullScript.commonPipelineEnvironment.configuration = [
            general: [buildTool: 'maven'],
            runStep: ['Pull-Request Voting': [npmExecuteScripts: true]]
        ]

        jsr.step.piperPipelineStagePRVoting(
            script: nullScript,
            juStabUtils: utils,
        )

        assertThat(stepsCalled, hasItems('npmExecuteScripts', 'testsPublishResults'))
    }
}
