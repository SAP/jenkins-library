#!groovy
import com.lesfurets.jenkins.unit.BasePipelineTest
import com.sap.piper.DefaultValueCache
import com.sap.piper.GitUtils
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.JenkinsLoggingRule
import util.JenkinsReadMavenPomRule
import util.JenkinsShellCallRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.junit.Assert.assertEquals

class ArtifactSetVersionTest extends BasePipelineTest {

    Script artifactSetVersionScript

    def cpe
    def gitUtils
    def sshAgentList = []

    ExpectedException thrown = ExpectedException.none()
    JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    JenkinsWriteFileRule jwfr = new JenkinsWriteFileRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jlr)
        .around(jscr)
        .around(new JenkinsReadMavenPomRule(this, 'test/resources/MavenArtifactVersioning'))
        .around(jwfr)

    @Before
    void init() throws Throwable {

        helper.registerAllowedMethod("sshagent", [List.class, Closure.class], { list, closure ->
            sshAgentList = list
            return closure()
        })

        jscr.setReturnValue('git rev-parse HEAD', 'testCommitId')
        jscr.setReturnValue("date +'%Y%m%d%H%M%S'", '20180101010203')
        jscr.setReturnValue('git diff --quiet HEAD', value: 0)

        cpe = loadScript('commonPipelineEnvironment.groovy').commonPipelineEnvironment
        artifactSetVersionScript = loadScript("artifactSetVersion.groovy")

        gitUtils = new GitUtils()
        prepareObjectInterceptors(gitUtils)
    }

    @Test
    void testVersioning() {
        artifactSetVersionScript.call(script: [commonPipelineEnvironment: cpe], juStabGitUtils: gitUtils, buildTool: 'maven', gitSshUrl: 'myGitSshUrl')

        assertEquals('1.2.3-20180101010203_testCommitId', cpe.getArtifactVersion())
        assertEquals('testCommitId', cpe.getGitCommitId())

        assertEquals('mvn versions:set -DnewVersion=1.2.3-20180101010203_testCommitId --file pom.xml', jscr.shell[3])
        assertEquals('git add .', jscr.shell[4])
        assertEquals ("git commit -m 'update version 1.2.3-20180101010203_testCommitId'", jscr.shell[5])
        assertEquals ("git remote set-url origin myGitSshUrl", jscr.shell[6])
        assertEquals ("git tag build_1.2.3-20180101010203_testCommitId", jscr.shell[7])
        assertEquals ("git push origin build_1.2.3-20180101010203_testCommitId", jscr.shell[8])
    }

    @Test
    void testVersioningCustomGitUserAndEMail() {
        artifactSetVersionScript.call(script: [commonPipelineEnvironment: cpe], juStabGitUtils: gitUtils, buildTool: 'maven', gitSshUrl: 'myGitSshUrl', gitUserEMail: 'test@test.com', gitUserName: 'test')

        assertEquals ('git -c user.email="test@test.com" -c user.name "test" commit -m \'update version 1.2.3-20180101010203_testCommitId\'', jscr.shell[5])
    }

    @Test
    void testVersioningWithTimestamp() {
        artifactSetVersionScript.call(script: [commonPipelineEnvironment: cpe], juStabGitUtils: gitUtils, buildTool: 'maven', timestamp: '2018')
        assertEquals('1.2.3-2018_testCommitId', cpe.getArtifactVersion())
    }

    @Test
    void testVersioningNoBuildTool() {
        thrown.expect(Exception)
        thrown.expectMessage('ERROR - NO VALUE AVAILABLE FOR buildTool')
        artifactSetVersionScript.call(script: [commonPipelineEnvironment: cpe], juStabGitUtils: gitUtils)
    }

    @Test
    void testVersioningWithCustomTemplate() {
        artifactSetVersionScript.call(script: [commonPipelineEnvironment: cpe], juStabGitUtils: gitUtils, buildTool: 'maven', versioningTemplate: '${version}-xyz')
        assertEquals('1.2.3-xyz', cpe.getArtifactVersion())
    }

    @Test
    void testVersioningWithTypeAppContainer() {
        cpe.setArtifactVersion('1.2.3-xyz')
        artifactSetVersionScript.call(script: [commonPipelineEnvironment: cpe], juStabGitUtils: gitUtils, buildTool: 'docker', artifactType: 'appContainer', dockerVersionSource: 'appVersion')
        assertEquals('1.2.3-xyz', cpe.getArtifactVersion())
        assertEquals('1.2.3-xyz', jwfr.files['VERSION'])
    }

    void prepareObjectInterceptors(object) {
        object.metaClass.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.static.invokeMethod = helper.getMethodInterceptor()
        object.metaClass.methodMissing = helper.getMethodMissingInterceptor()
    }


}
