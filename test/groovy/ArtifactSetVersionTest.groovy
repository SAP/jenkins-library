#!groovy
import com.lesfurets.jenkins.unit.BasePipelineTest
import com.sap.piper.GitUtils
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.junit.Assert.assertEquals

class ArtifactSetVersionTest extends BasePipelineTest {
    def gitUtils
    def sshAgentList = []

    ExpectedException thrown = ExpectedException.none()
    JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    JenkinsWriteFileRule jwfr = new JenkinsWriteFileRule(this)
    JenkinsStepRule jsr = new JenkinsStepRule(this)
    JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jlr)
        .around(jscr)
        .around(new JenkinsReadMavenPomRule(this, 'test/resources/MavenArtifactVersioning'))
        .around(jwfr)
        .around(jsr)
        .around(jer)

    @Before
    void init() throws Throwable {
        helper.registerAllowedMethod("sshagent", [List.class, Closure.class], { list, closure ->
            sshAgentList = list
            return closure()
        })

        jscr.setReturnValue('git rev-parse HEAD', 'testCommitId')
        jscr.setReturnValue("date +'%Y%m%d%H%M%S'", '20180101010203')
        jscr.setReturnValue('git diff --quiet HEAD', 0)

        gitUtils = new GitUtils()
        prepareObjectInterceptors(gitUtils)
    }

    @Test
    void testVersioning() {
        jsr.step.call(script: [commonPipelineEnvironment: jer.env], juStabGitUtils: gitUtils, buildTool: 'maven', gitSshUrl: 'myGitSshUrl')

        assertEquals('1.2.3-20180101010203_testCommitId', jer.env.getArtifactVersion())
        assertEquals('testCommitId', jer.env.getGitCommitId())

        assertEquals('mvn versions:set -DnewVersion=1.2.3-20180101010203_testCommitId --file pom.xml', jscr.shell[3])
        assertEquals('git add .', jscr.shell[4])
        assertEquals ("git commit -m 'update version 1.2.3-20180101010203_testCommitId'", jscr.shell[5])
        assertEquals ("git remote set-url origin myGitSshUrl", jscr.shell[6])
        assertEquals ("git tag build_1.2.3-20180101010203_testCommitId", jscr.shell[7])
        assertEquals ("git push origin build_1.2.3-20180101010203_testCommitId", jscr.shell[8])
    }

    @Test
    void testVersioningWithoutCommit() {
        jsr.step.call(script: [commonPipelineEnvironment: jer.env], juStabGitUtils: gitUtils, buildTool: 'maven', commitVersion: false)

        assertEquals('1.2.3-20180101010203_testCommitId', jer.env.getArtifactVersion())
        assertEquals('mvn versions:set -DnewVersion=1.2.3-20180101010203_testCommitId --file pom.xml', jscr.shell[3])
    }

    @Test
    void testVersioningCustomGitUserAndEMail() {
        jsr.step.call(script: [commonPipelineEnvironment: jer.env], juStabGitUtils: gitUtils, buildTool: 'maven', gitSshUrl: 'myGitSshUrl', gitUserEMail: 'test@test.com', gitUserName: 'test')

        assertEquals ('git -c user.email="test@test.com" -c user.name "test" commit -m \'update version 1.2.3-20180101010203_testCommitId\'', jscr.shell[5])
    }

    @Test
    void testVersioningWithTimestamp() {
        jsr.step.call(script: [commonPipelineEnvironment: jer.env], juStabGitUtils: gitUtils, buildTool: 'maven', timestamp: '2018')
        assertEquals('1.2.3-2018_testCommitId', jer.env.getArtifactVersion())
    }

    @Test
    void testVersioningNoBuildTool() {
        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR buildTool')
        jsr.step.call(script: [commonPipelineEnvironment: jer.env], juStabGitUtils: gitUtils)
    }

    @Test
    void testVersioningWithCustomTemplate() {
        jsr.step.call(script: [commonPipelineEnvironment: jer.env], juStabGitUtils: gitUtils, buildTool: 'maven', versioningTemplate: '${version}-xyz')
        assertEquals('1.2.3-xyz', jer.env.getArtifactVersion())
    }

    @Test
    void testVersioningWithTypeAppContainer() {
        jer.env.setArtifactVersion('1.2.3-xyz')
        jsr.step.call(script: [commonPipelineEnvironment: jer.env], juStabGitUtils: gitUtils, buildTool: 'docker', artifactType: 'appContainer', dockerVersionSource: 'appVersion')
        assertEquals('1.2.3-xyz', jer.env.getArtifactVersion())
        assertEquals('1.2.3-xyz', jwfr.files['VERSION'])
    }

    void prepareObjectInterceptors(object) {
        object.metaClass.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.static.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.methodMissing = helper.getMethodMissingInterceptor()
    }


}
