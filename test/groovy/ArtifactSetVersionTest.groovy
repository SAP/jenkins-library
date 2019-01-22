#!groovy

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.GitUtils

import hudson.AbortException
import util.BasePiperTest
import util.JenkinsDockerExecuteRule
import util.JenkinsEnvironmentRule
import util.JenkinsLoggingRule
import util.JenkinsReadMavenPomRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.hasItems
import static org.hamcrest.Matchers.not
import static org.hamcrest.Matchers.notNullValue
import static org.hamcrest.Matchers.stringContainsInOrder
import static org.hamcrest.Matchers.containsString
import static org.junit.Assert.assertThat

import static org.junit.Assert.assertEquals

class ArtifactSetVersionTest extends BasePiperTest {
    Map dockerParameters

    def GitUtils gitUtils = new GitUtils() {
        boolean isWorkTreeDirty() {
            return false
        }

        String getGitCommitIdOrNull() {
            return 'testCommitId'
        }
    }

    def sshAgentList = []

    private ExpectedException thrown = ExpectedException.none()
    private JenkinsDockerExecuteRule jder = new JenkinsDockerExecuteRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsWriteFileRule jwfr = new JenkinsWriteFileRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(loggingRule)
        .around(shellRule)
        .around(new JenkinsReadMavenPomRule(this, 'test/resources/versioning/MavenArtifactVersioning'))
        .around(jwfr)
        .around(jder)
        .around(stepRule)
        .around(jer)

    @Before
    void init() throws Throwable {
        dockerParameters = [:]

        nullScript.commonPipelineEnvironment.setArtifactVersion(null)
        nullScript.commonPipelineEnvironment.setGitSshUrl('git@test.url')

        helper.registerAllowedMethod("sshagent", [List.class, Closure.class], { list, closure ->
            sshAgentList = list
            return closure()
        })

        shellRule.setReturnValue("date --universal +'%Y%m%d%H%M%S'", '20180101010203')
        shellRule.setReturnValue('git diff --quiet HEAD', 0)

        helper.registerAllowedMethod('fileExists', [String.class], {true})
    }

    @Test
    void testVersioning() {
        stepRule.step.artifactSetVersion(script: stepRule.step, juStabGitUtils: gitUtils, buildTool: 'maven', gitSshUrl: 'myGitSshUrl')

        assertEquals('1.2.3-20180101010203_testCommitId', jer.env.getArtifactVersion())
        assertEquals('testCommitId', jer.env.getGitCommitId())

        assertThat(shellRule.shell, hasItem("mvn --file 'pom.xml' --batch-mode -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn versions:set -DnewVersion=1.2.3-20180101010203_testCommitId -DgenerateBackupPoms=false"))
        assertThat(shellRule.shell.join(), stringContainsInOrder([
                                            "git add .",
                                            "git commit -m 'update version 1.2.3-20180101010203_testCommitId'",
                                            'git tag build_1.2.3-20180101010203_testCommitId',
                                            'git push myGitSshUrl build_1.2.3-20180101010203_testCommitId',
                                            ]
                                        ))
    }

    @Test
    void testVersioningWithoutCommit() {
        stepRule.step.artifactSetVersion(script: stepRule.step, juStabGitUtils: gitUtils, buildTool: 'maven', commitVersion: false)

        assertEquals('1.2.3-20180101010203_testCommitId', jer.env.getArtifactVersion())
        assertThat(shellRule.shell, hasItem("mvn --file 'pom.xml' --batch-mode -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn versions:set -DnewVersion=1.2.3-20180101010203_testCommitId -DgenerateBackupPoms=false"))
        assertThat(shellRule.shell, not(hasItem(containsString('commit'))))
    }

    @Test
    void testVersioningCustomGitUserAndEMail() {
        stepRule.step.artifactSetVersion(script: stepRule.step, juStabGitUtils: gitUtils, buildTool: 'maven', gitSshUrl: 'myGitSshUrl', gitUserEMail: 'test@test.com', gitUserName: 'test')

        assertThat(shellRule.shell, hasItem(containsString("git -c user.email=\"test@test.com\" -c user.name=\"test\" commit -m 'update version 1.2.3-20180101010203_testCommitId'")))
    }

    @Test
    void testVersioningWithTimestamp() {
        stepRule.step.artifactSetVersion(script: stepRule.step, juStabGitUtils: gitUtils, buildTool: 'maven', timestamp: '2018')
        assertEquals('1.2.3-2018_testCommitId', jer.env.getArtifactVersion())
    }

    @Test
    void testVersioningNoBuildTool() {
        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR buildTool')
        stepRule.step.artifactSetVersion(script: stepRule.step, juStabGitUtils: gitUtils)
    }

    @Test
    void testVersioningWithCustomTemplate() {
        stepRule.step.artifactSetVersion(script: stepRule.step, juStabGitUtils: gitUtils, buildTool: 'maven', versioningTemplate: '${version}-xyz')
        assertEquals('1.2.3-xyz', jer.env.getArtifactVersion())
    }

    @Test
    void testVersioningWithTypeAppContainer() {
        nullScript.commonPipelineEnvironment.setAppContainerProperty('gitSshUrl', 'git@test.url')
        jer.env.setArtifactVersion('1.2.3-xyz')
        stepRule.step.artifactSetVersion(script: stepRule.step, juStabGitUtils: gitUtils, buildTool: 'docker', artifactType: 'appContainer', dockerVersionSource: 'appVersion')
        assertEquals('1.2.3-xyz', jer.env.getArtifactVersion())
        assertEquals('1.2.3-xyz', jwfr.files['VERSION'])
    }

    @Test
    void testCredentialCompatibility() {
        stepRule.step.artifactSetVersion (
            script: nullScript,
            buildTool: 'maven',
            gitCredentialsId: 'testCredentials',
            juStabGitUtils: gitUtils
        )
        assertThat(sshAgentList, hasItem('testCredentials'))
    }
}
