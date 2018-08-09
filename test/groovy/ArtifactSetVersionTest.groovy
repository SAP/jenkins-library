#!groovy
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsDockerExecuteRule
import util.JenkinsEnvironmentRule
import util.JenkinsLoggingRule
import util.JenkinsReadMavenPomRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.PluginMock
import util.Rules

import static org.hamcrest.Matchers.hasItem
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertEquals

class ArtifactSetVersionTest extends BasePiperTest {
    Map dockerParameters

    def sshAgentList = []

    private ExpectedException thrown = ExpectedException.none()
    private JenkinsDockerExecuteRule jder = new JenkinsDockerExecuteRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsWriteFileRule jwfr = new JenkinsWriteFileRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jlr)
        .around(jscr)
        .around(new JenkinsReadMavenPomRule(this, 'test/resources/MavenArtifactVersioning'))
        .around(jwfr)
        .around(jder)
        .around(jsr)
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

        jscr.setReturnValue('git rev-parse HEAD', 'testCommitId')
        jscr.setReturnValue("date --universal +'%Y%m%d%H%M%S'", '20180101010203')
        jscr.setReturnValue('git diff --quiet HEAD', 0)
        jscr.setReturnValue('git rev-parse --is-inside-work-tree 1>/dev/null 2>&1', 0)

        binding.setVariable('Jenkins', [instance: [pluginManager: [plugins: [new PluginMock('docker-workflow')]]]])

        helper.registerAllowedMethod('fileExists', [String.class], {true})
    }

    @Test
    void testVersioning() {
        jsr.step.artifactSetVersion(script: jsr.step, juStabGitUtils: gitUtils, buildTool: 'maven', gitSshUrl: 'myGitSshUrl')

        assertEquals('1.2.3-20180101010203_testCommitId', jer.env.getArtifactVersion())
        assertEquals('testCommitId', jer.env.getGitCommitId())

        assertThat(jscr.shell, hasItem("mvn --file 'pom.xml' --batch-mode -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn versions:set -DnewVersion=1.2.3-20180101010203_testCommitId"))
        assertThat(jscr.shell, hasItem('git add .'))
        assertThat(jscr.shell, hasItem("git commit -m 'update version 1.2.3-20180101010203_testCommitId'"))
        assertThat(jscr.shell, hasItem("git remote set-url origin myGitSshUrl"))
        assertThat(jscr.shell, hasItem("git tag build_1.2.3-20180101010203_testCommitId"))
        assertThat(jscr.shell, hasItem("git push origin build_1.2.3-20180101010203_testCommitId"))
    }

    @Test
    void testVersioningWithoutCommit() {
        jsr.step.artifactSetVersion(script: jsr.step, juStabGitUtils: gitUtils, buildTool: 'maven', commitVersion: false)

        assertEquals('1.2.3-20180101010203_testCommitId', jer.env.getArtifactVersion())
        assertThat(jscr.shell, hasItem("mvn --file 'pom.xml' --batch-mode -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn versions:set -DnewVersion=1.2.3-20180101010203_testCommitId"))
    }

    @Test
    void testVersioningWithoutScript() {
        jsr.step.artifactSetVersion(juStabGitUtils: gitUtils, buildTool: 'maven', commitVersion: false)

        assertEquals('1.2.3-20180101010203_testCommitId', jer.env.getArtifactVersion())
        assertThat(jscr.shell, hasItem("mvn --file 'pom.xml' --batch-mode -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn versions:set -DnewVersion=1.2.3-20180101010203_testCommitId"))
    }

    @Test
    void testVersioningCustomGitUserAndEMail() {
        jsr.step.artifactSetVersion(script: jsr.step, juStabGitUtils: gitUtils, buildTool: 'maven', gitSshUrl: 'myGitSshUrl', gitUserEMail: 'test@test.com', gitUserName: 'test')

        assertThat(jscr.shell, hasItem("git -c user.email=\"test@test.com\" -c user.name=\"test\" commit -m 'update version 1.2.3-20180101010203_testCommitId'"))
    }

    @Test
    void testVersioningWithTimestamp() {
        jsr.step.artifactSetVersion(script: jsr.step, juStabGitUtils: gitUtils, buildTool: 'maven', timestamp: '2018')
        assertEquals('1.2.3-2018_testCommitId', jer.env.getArtifactVersion())
    }

    @Test
    void testVersioningNoBuildTool() {
        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR buildTool')
        jsr.step.artifactSetVersion(script: jsr.step, juStabGitUtils: gitUtils)
    }

    @Test
    void testVersioningWithCustomTemplate() {
        jsr.step.artifactSetVersion(script: jsr.step, juStabGitUtils: gitUtils, buildTool: 'maven', versioningTemplate: '${version}-xyz')
        assertEquals('1.2.3-xyz', jer.env.getArtifactVersion())
    }

    @Test
    void testVersioningWithTypeAppContainer() {
        nullScript.commonPipelineEnvironment.setAppContainerProperty('gitSshUrl', 'git@test.url')
        jer.env.setArtifactVersion('1.2.3-xyz')
        jsr.step.artifactSetVersion(script: jsr.step, juStabGitUtils: gitUtils, buildTool: 'docker', artifactType: 'appContainer', dockerVersionSource: 'appVersion')
        assertEquals('1.2.3-xyz', jer.env.getArtifactVersion())
        assertEquals('1.2.3-xyz', jwfr.files['VERSION'])
    }

    @Test
    void testCredentialCompatibility() {
        jsr.step.artifactSetVersion (
            script: nullScript,
            buildTool: 'maven',
            gitCredentialsId: 'testCredentials',
            juStabGitUtils: gitUtils
        )
        assertThat(sshAgentList, hasItem('testCredentials'))
    }

    void prepareObjectInterceptors(object) {
        object.metaClass.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.static.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.methodMissing = helper.getMethodMissingInterceptor()
    }

}
