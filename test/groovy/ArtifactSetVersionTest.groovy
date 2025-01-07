import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.GitUtils
import com.sap.piper.Utils

import hudson.AbortException
import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsDockerExecuteRule
import util.JenkinsEnvironmentRule
import util.JenkinsLoggingRule
import util.JenkinsMavenExecuteRule
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
import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.emptyIterable
import static org.junit.Assert.assertThat

import org.hamcrest.Matchers

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
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsMavenExecuteRule mvnExecuteRule = new JenkinsMavenExecuteRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsEnvironmentRule environmentRule = new JenkinsEnvironmentRule(this)
    private JenkinsCredentialsRule jenkinsCredentialsRule = new JenkinsCredentialsRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(loggingRule)
        .around(shellRule)
        .around(new JenkinsReadMavenPomRule(this, 'test/resources/versioning/MavenArtifactVersioning'))
        .around(writeFileRule)
        .around(dockerExecuteRule)
        .around(stepRule)
        .around(jenkinsCredentialsRule)
        .around(environmentRule)
        .around(mvnExecuteRule)

    @Before
    void init() throws Throwable {
        dockerParameters = [:]
        String version = '1.2.3'

        nullScript.commonPipelineEnvironment.setArtifactVersion(null)
        nullScript.commonPipelineEnvironment.setGitSshUrl('git@test.url')

        helper.registerAllowedMethod("sshagent", [List.class, Closure.class], { list, closure ->
            sshAgentList = list
            return closure()
        })

        mvnExecuteRule.setReturnValue([
            'pomPath': 'pom.xml',
            'goals': ['org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate'],
            'defines': ['-Dexpression=project.version', '-DforceStdout', '-q'],
        ], version)

        mvnExecuteRule.setReturnValue([
            'pomPath': 'snapshot/pom.xml',
            'goals': ['org.apache.maven.plugins:maven-help-plugin:3.1.0:evaluate'],
            'defines': ['-Dexpression=project.version', '-DforceStdout', '-q'],
        ], version)

        shellRule.setReturnValue("date --utc +'%Y%m%d%H%M%S'", '20180101010203')
        shellRule.setReturnValue('git diff --quiet HEAD', 0)

        helper.registerAllowedMethod('fileExists', [String.class], {true})

        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
    }

    @Test
    void testVersioningPushViaSSH() {
        stepRule.step.artifactSetVersion(script: stepRule.step, juStabGitUtils: gitUtils, buildTool: 'maven', gitSshUrl: 'myGitSshUrl')

        assertEquals('1.2.3-20180101010203_testCommitId', environmentRule.env.getArtifactVersion())
        assertEquals('testCommitId', environmentRule.env.getGitCommitId())

        assertEquals(new JenkinsMavenExecuteRule.Execution([
            pomPath: 'pom.xml',
            goals: ['org.codehaus.mojo:versions-maven-plugin:2.7:set'],
            defines: ['-DnewVersion=1.2.3-20180101010203_testCommitId', '-DgenerateBackupPoms=false']
        ]), mvnExecuteRule.executions[1])

        assertThat(shellRule.shell.join(), stringContainsInOrder([
            "git add .",
            "git commit -m 'update version 1.2.3-20180101010203_testCommitId'",
            "git tag 'build_1.2.3-20180101010203_testCommitId'",
            "git push 'myGitSshUrl' 'build_1.2.3-20180101010203_testCommitId'",
            ]
        ))
    }

    @Test
    void testVersioningNoPush() {

        stepRule.step.artifactSetVersion(
            script: stepRule.step,
            juStabGitUtils: gitUtils,
            buildTool: 'maven',
            gitPushMode: 'NONE')

        assertThat(loggingRule.log, containsString('Git push to remote has been skipped.'))
        assertThat(((Iterable)shellRule.shell).join(), not(containsString('push')))
    }

    @Test
    void testVersioningPushViaHTTPS() {

        jenkinsCredentialsRule.withCredentials('myGitRepoCredentials', 'me', 'topSecret')

        stepRule.step.artifactSetVersion(
            script: stepRule.step,
            juStabGitUtils: gitUtils,
            buildTool: 'maven',
            gitHttpsCredentialsId: 'myGitRepoCredentials',
            gitHttpsUrl: 'https://example.org/myGitRepo',
            gitPushMode: 'HTTPS')

        // closer version checks already performed in test 'testVersioningPushViaSSH', focusing on
        // GIT related assertions here

        assertEquals(new JenkinsMavenExecuteRule.Execution([
            pomPath: 'pom.xml',
            goals: ['org.codehaus.mojo:versions-maven-plugin:2.7:set'],
            defines: ['-DnewVersion=1.2.3-20180101010203_testCommitId', '-DgenerateBackupPoms=false']
        ]), mvnExecuteRule.executions[1])
        assertThat(((Iterable)shellRule.shell).join(), stringContainsInOrder([
            "git add .",
            "git commit -m 'update version 1.2.3-20180101010203_testCommitId'",
            "git tag 'build_1.2.3-20180101010203_testCommitId'",
            "git push https://me:topSecret@example.org/myGitRepo 'build_1.2.3-20180101010203_testCommitId'",
            ]
        ))
    }

    @Test
    void testVersioningPushViaHTTPDisableSSLCheck() {

        jenkinsCredentialsRule.withCredentials('myGitRepoCredentials', 'me', 'topSecret')

        stepRule.step.artifactSetVersion(
            script: stepRule.step,
            juStabGitUtils: gitUtils,
            buildTool: 'maven',
            gitHttpsCredentialsId: 'myGitRepoCredentials',
            gitHttpsUrl: 'https://example.org/myGitRepo',
            gitPushMode: 'HTTPS',
            gitDisableSslVerification: true)

        // closer version checks already performed in test 'testVersioningPushViaSSH', focusing on
        // GIT related assertions here

        assertThat(((Iterable)shellRule.shell).join(), containsString('-c http.sslVerify=false'))
    }

    @Test
    void testVersioningPushViaHTTPVerboseMode() {

        jenkinsCredentialsRule.withCredentials('myGitRepoCredentials', 'me', 'topSecret')

        stepRule.step.artifactSetVersion(
            script: stepRule.step,
            juStabGitUtils: gitUtils,
            buildTool: 'maven',
            gitHttpsCredentialsId: 'myGitRepoCredentials',
            gitHttpsUrl: 'https://example.org/myGitRepo',
            gitPushMode: 'HTTPS',
            verbose: true)

        // closer version checks already performed in test 'testVersioningPushViaSSH', focusing on
        // GIT related assertions here

        assertThat(((Iterable)shellRule.shell).join(), allOf(
            containsString('GIT_CURL_VERBOSE=1'),
            containsString('GIT_TRACE=1'),
            containsString('--verbose'),
            not(containsString('&>/dev/null'))))
    }

    @Test
    void testVersioningPushViaHTTPSInDebugModeEncodingDoesNotRevealSecrets() {

        loggingRule.expect('Verbose flag set, but encoded username/password differs from unencoded version. Cannot provide verbose output in this case.')
        loggingRule.expect('Performing git push in quiet mode')

        jenkinsCredentialsRule.withCredentials('myGitRepoCredentials', 'me', 'top@Secret')

        stepRule.step.artifactSetVersion(
            script: stepRule.step,
            juStabGitUtils: gitUtils,
            buildTool: 'maven',
            gitHttpsCredentialsId: 'myGitRepoCredentials',
            gitHttpsUrl: 'https://example.org/myGitRepo',
            gitPushMode: 'HTTPS',
            verbose: true)

        // closer version checks already performed in test 'testVersioningPushViaSSH', focusing on
        // GIT related assertions here

        assertThat(((Iterable)shellRule.shell).join(), stringContainsInOrder([
            "git add .",
            "git commit -m 'update version 1.2.3-20180101010203_testCommitId'",
            "git tag 'build_1.2.3-20180101010203_testCommitId'",
            "#!/bin/bash -e git push --quiet https://me:top%40Secret@example.org/myGitRepo 'build_1.2.3-20180101010203_testCommitId' &>/dev/null",
            ]
        ))
    }


    @Test
    void testVersioningPushViaHTTPSEncodingDoesNotRevealSecrets() {

        // Credentials needs to be url encoded. In case that changes the secrets the credentials plugin
        // doesn't hide the secrets anymore in the log. Hence we have to take care that the command is silent.
        // Check for more details how that is handled in the step.

        loggingRule.expect('Performing git push in quiet mode')

        jenkinsCredentialsRule.withCredentials('myGitRepoCredentials', 'me', 'top@Secret')

        stepRule.step.artifactSetVersion(
            script: stepRule.step,
            juStabGitUtils: gitUtils,
            buildTool: 'maven',
            gitHttpsCredentialsId: 'myGitRepoCredentials',
            gitHttpsUrl: 'https://example.org/myGitRepo',
            gitPushMode: 'HTTPS')

        // closer version checks already performed in test 'testVersioningPushViaSSH', focusing on
        // GIT related assertions here

        assertThat(((Iterable)shellRule.shell).join(), stringContainsInOrder([
            "git add .",
            "git commit -m 'update version 1.2.3-20180101010203_testCommitId'",
            "git tag 'build_1.2.3-20180101010203_testCommitId'",
            "#!/bin/bash -e git push --quiet https://me:top%40Secret@example.org/myGitRepo 'build_1.2.3-20180101010203_testCommitId' &>/dev/null",
            ]
        ))
    }

    @Test
    void testVersioningWithoutCommit() {
        stepRule.step.artifactSetVersion(script: stepRule.step, juStabGitUtils: gitUtils, buildTool: 'maven', commitVersion: false)

        assertEquals('1.2.3-20180101010203_testCommitId', environmentRule.env.getArtifactVersion())
        assertEquals(new JenkinsMavenExecuteRule.Execution([
            pomPath: 'pom.xml',
            goals: ['org.codehaus.mojo:versions-maven-plugin:2.7:set'],
            defines: ['-DnewVersion=1.2.3-20180101010203_testCommitId', '-DgenerateBackupPoms=false']
        ]), mvnExecuteRule.executions[1])
        assertThat(shellRule.shell, not(hasItem(containsString('commit'))))
    }

    @Test
    void testVersioningCustomGitUserAndEMail() {
        stepRule.step.artifactSetVersion(script: stepRule.step, juStabGitUtils: gitUtils, buildTool: 'maven', gitSshUrl: 'myGitSshUrl', gitUserEMail: 'test@test.com', gitUserName: 'test')

        assertThat(shellRule.shell, hasItem(containsString("git -c user.email='test@test.com' -c user.name='test' commit -m 'update version 1.2.3-20180101010203_testCommitId'")))
    }

    @Test
    void testVersioningWithTimestamp() {
        stepRule.step.artifactSetVersion(script: stepRule.step, juStabGitUtils: gitUtils, buildTool: 'maven', timestamp: '2018')
        assertEquals('1.2.3-2018_testCommitId', environmentRule.env.getArtifactVersion())
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
        assertEquals('1.2.3-xyz', environmentRule.env.getArtifactVersion())
    }

    @Test
    void testVersioningWithTypeAppContainer() {
        nullScript.commonPipelineEnvironment.setAppContainerProperty('gitSshUrl', 'git@test.url')
        environmentRule.env.setArtifactVersion('1.2.3-xyz')
        stepRule.step.artifactSetVersion(script: stepRule.step, juStabGitUtils: gitUtils, buildTool: 'docker', artifactType: 'appContainer', dockerVersionSource: 'appVersion')
        assertEquals('1.2.3-xyz', environmentRule.env.getArtifactVersion())
        assertEquals('1.2.3-xyz', writeFileRule.files['VERSION'])
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
