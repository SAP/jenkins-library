import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*
import com.sap.piper.Utils

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class MailSendNotificationTest extends BasePiperTest {
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(loggingRule)
        .around(shellRule)
        .around(stepRule)

    @Before
    void init() throws Exception {
        // register Jenkins commands with mock values
        helper.registerAllowedMethod("deleteDir", [], null)
        helper.registerAllowedMethod("sshagent", [Map.class, Closure.class], null)

        nullScript.commonPipelineEnvironment.configuration = nullScript.commonPipelineEnvironment.configuration ?: [:]
        nullScript.commonPipelineEnvironment.configuration['general'] = nullScript.commonPipelineEnvironment.configuration['general'] ?: [:]
        nullScript.commonPipelineEnvironment.configuration['steps'] = nullScript.commonPipelineEnvironment.configuration['steps'] ?: [:]
        nullScript.commonPipelineEnvironment.configuration['steps']['mailSendNotification'] = nullScript.commonPipelineEnvironment.configuration['steps']['mailSendNotification'] ?: [:]

        helper.registerAllowedMethod('requestor', [], { -> return [$class: 'RequesterRecipientProvider']})

        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
    }

    @Test
    void testGetDistinctRecipients() throws Exception {
        // git log -10 --pretty=format:"%ae %ce"
        def input = '''user1@domain.com noreply+github@domain.com
user1@domain.com noreply+github@domain.com
user3@domain.com noreply+github@domain.com
user2@domain.com user1@domain.com
user1@noreply.domain.com
user1@domain.com noreply+github@domain.com
user3@domain.com noreply+github@domain.com
user3@domain.com noreply+github@domain.com
user3@domain.com noreply+github@domain.com
user3@domain.com noreply+github@domain.com
user3@domain.com noreply+github@domain.com'''

        def result = stepRule.step.getDistinctRecipients(input)
        // asserts
        assertThat(result.split(' '), arrayWithSize(3))
        assertThat(result, containsString('user1@domain.com'))
        assertThat(result, containsString('user2@domain.com'))
        assertThat(result, containsString('user3@domain.com'))
    }

    @Test
    void testCulpritsFromGitCommit() throws Exception {
        def gitCommand = "git log -2 --first-parent --pretty=format:'%ae %ce'"
        def expected = "user2@domain.com user3@domain.com"

        shellRule.setReturnValue("git log -2 --first-parent --pretty=format:'%ae %ce'", 'user2@domain.com user3@domain.com')

        def result = stepRule.step.getCulprits(
            [
                gitSSHCredentialsId: '',
                gitUrl: 'git@github.domain.com:IndustryCloudFoundation/pipeline-test-node.git',
                gitCommitId: 'f0973368a35a2b973612acb86f932c61f2635f6e'
            ],
            'master',
            2)
        // asserts
        assertThat(result, containsString('user2@domain.com'))
        assertThat(result, containsString('user3@domain.com'))
    }

    @Test
    void testCulpritsWithEmptyGitCommit() throws Exception {

        shellRule.setReturnValue('git log > /dev/null 2>&1',1)

        stepRule.step.getCulprits(
            [
                gitSSHCredentialsId: '',
                gitUrl: 'git@github.domain.com:IndustryCloudFoundation/pipeline-test-node.git',
                gitCommitId: ''
            ],
            'master',
            2)
        // asserts
        assertThat(loggingRule.log, containsString('[mailSendNotification] No git context available to retrieve culprits'))
    }

    @Test
    void testCulpritsWithoutGitCommit() throws Exception {

        shellRule.setReturnValue('git log > /dev/null 2>&1',1)

        stepRule.step.getCulprits(
            [
                gitSSHCredentialsId: '',
                gitUrl: 'git@github.domain.com:IndustryCloudFoundation/pipeline-test-node.git',
                gitCommitId: null
            ],
            'master',
            2)
        // asserts
        assertThat(loggingRule.log, containsString('[mailSendNotification] No git context available to retrieve culprits'))
    }

    @Test
    void testCulpritsWithoutBranch() throws Exception {

        shellRule.setReturnValue('git log > /dev/null 2>&1',1)

        stepRule.step.getCulprits(
            [
                gitSSHCredentialsId: '',
                gitUrl: 'git@github.domain.com:IndustryCloudFoundation/pipeline-test-node.git',
                gitCommitId: ''
            ],
            null,
            2)
        // asserts
        assertThat(loggingRule.log, containsString('[mailSendNotification] No git context available to retrieve culprits'))
    }

    @Test
    void testSendNotificationMail() throws Exception {
        def emailParameters = [:]
        def buildMock = [
            fullProjectName: 'testProjectName',
            displayName: 'testDisplayName',
            result: 'FAILURE',
            rawBuild: [
                getLog: { cnt -> return ['Setting http proxy: proxy.domain.com:8080',
' > git fetch --no-tags --progress https://github.com/SAP/jenkins-library.git +refs/heads/*:refs/remotes/origin/*',
'Checking out Revision myUniqueCommitId (master)',
' > git config core.sparsecheckout # timeout=10',
' > git checkout -f myUniqueCommitId',
'Commit message: "Merge pull request #147 from marcusholl/pr/useGitRevParseForInsideGitRepoCheck"',
' > git rev-list --no-walk myUniqueCommitId # timeout=10',
'[Pipeline] node',
'Running on Jenkins in /var/jenkins_home/workspace/Test/UserId/ECHO',
'[Pipeline] {',
'[Pipeline] stage',
'[Pipeline] { (A)',
'[Pipeline] script',
'[Pipeline] {']
                }
            ],
            getPreviousBuild: {
                return null
            }
        ]
        nullScript.currentBuild = buildMock
        nullScript.commonPipelineEnvironment.configuration['steps']['mailSendNotification']['notificationRecipients'] = 'piper@domain.com'
        helper.registerAllowedMethod('emailext', [Map.class], { map ->
            emailParameters = map
            return ''
        })

        stepRule.step.mailSendNotification(
            script: nullScript,
            notifyCulprits: false,
            gitUrl: 'git@github.domain.com:IndustryCloudFoundation/pipeline-test-node.git'
        )
        // asserts
        assertThat(emailParameters.to, is('piper@domain.com'))
        assertThat(emailParameters.subject, is('FAILURE: Build testProjectName testDisplayName'))
        assertThat(emailParameters.body, startsWith('<a href="http://build.url">http://build.url</a>\n<br>\nTo have a detailed look at the different pipeline stages: <a href="null">null</a>\n<br>\n<h3>Last lines of output</h3>'))
        assertThat(emailParameters.body, containsString(' > git fetch --no-tags --progress https://github.com/SAP/jenkins-library.git +refs/heads/*:refs/remotes/origin/*'))
        assertJobStatusSuccess()
    }

    @Test
    void testSendNotificationMailWithGeneralConfig() throws Exception {
        def credentials
        nullScript.currentBuild = [
            fullProjectName: 'testProjectName',
            displayName: 'testDisplayName',
            result: 'FAILURE',
            rawBuild: [ getLog: { cnt -> return ['empty'] } ],
            getChangeSets: { return null },
            getPreviousBuild: { return null }
        ]
        nullScript.commonPipelineEnvironment.configuration['general']['gitSshKeyCredentialsId'] = 'myCredentialsId'
        helper.registerAllowedMethod('emailext', [Map.class], null)
        helper.registerAllowedMethod("sshagent", [Map.class, Closure.class], { map, closure ->
            credentials = map.credentials
            return null
        })

        shellRule.setReturnValue("git log -0 --pretty=format:'%ae %ce'", 'user2@domain.com user3@domain.com')

        stepRule.step.mailSendNotification(
            script: nullScript,
            gitCommitId: 'abcd1234',
            //notifyCulprits: true,
            gitUrl: 'git@github.domain.com:IndustryCloudFoundation/pipeline-test-node.git'
        )
        // asserts
        assertThat(credentials, hasItem('myCredentialsId'))
        assertJobStatusSuccess()
    }

    @Test
    void testSendNotificationMailWithEmptySshKey() throws Exception {
        def credentials
        nullScript.currentBuild = [
            fullProjectName: 'testProjectName',
            displayName: 'testDisplayName',
            result: 'FAILURE',
            rawBuild: [ getLog: { cnt -> return ['empty'] } ],
            getChangeSets: { return null },
            getPreviousBuild: { return null }
        ]
        helper.registerAllowedMethod('emailext', [Map.class], null)
        helper.registerAllowedMethod("sshagent", [Map.class, Closure.class], { map, closure ->
            credentials = map.credentials
            return null
        })

        shellRule.setReturnValue("git log -0 --pretty=format:'%ae %ce'", 'user2@domain.com user3@domain.com')

        stepRule.step.mailSendNotification(
            script: nullScript,
            gitCommitId: 'abcd1234',
            gitUrl: 'git@github.domain.com:IndustryCloudFoundation/pipeline-test-node.git'
        )
        // asserts
        assertThat(credentials, hasItem(''))
        assertJobStatusSuccess()
    }

    @Test
    void testSendNotificationMailOnFirstBuild() throws Exception {
        def emailExtCalls = []
        def buildMock = [
            fullProjectName: 'testProjectName',
            displayName: 'testDisplayName',
            result: 'SUCCESS',
            getPreviousBuild: {
                return null
            }
        ]
        nullScript.currentBuild = buildMock
        helper.registerAllowedMethod('emailext', [Map.class], { map ->
            emailExtCalls.add(map)
            return ''
        })

        stepRule.step.mailSendNotification(
            script: nullScript,
            notifyCulprits: false,
            gitUrl: 'git@github.domain.com:IndustryCloudFoundation/pipeline-test-node.git'
        )

        assertThat(emailExtCalls, hasSize(0))
    }

    @Test
    void testSendNotificationMailOnRecovery() throws Exception {
        def emailExtCalls = []
        def buildMock = [
            fullProjectName: 'testProjectName',
            displayName: 'testDisplayName',
            result: 'SUCCESS',
            getPreviousBuild: {
                return [result: 'FAILURE']
            }
        ]
        nullScript.currentBuild = buildMock
        helper.registerAllowedMethod('emailext', [Map.class], { map ->
            emailExtCalls.add(map)
            return ''
        })

        stepRule.step.mailSendNotification(
            script: nullScript,
            notifyCulprits: false,
            gitUrl: 'git@github.domain.com:IndustryCloudFoundation/pipeline-test-node.git'
        )

        assertThat(emailExtCalls, hasSize(1))
        assertThat(emailExtCalls[0].subject, is("SUCCESS: Build testProjectName testDisplayName is back to normal"))
    }
}
