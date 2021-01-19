import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules
import com.sap.piper.Utils

import static org.junit.Assert.*

class SlackSendNotificationTest extends BasePiperTest {
    def slackCallMap = [:]

    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(loggingRule)
        .around(stepRule)

    @Before
    void init() throws Exception {
        helper.registerAllowedMethod("slackSend", [Map.class], {m -> slackCallMap = m})
        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
    }

    @Test
    void testNotificationBuildSuccessDefaultChannel() throws Exception {
        stepRule.step.slackSendNotification(script: [currentBuild: [result: 'SUCCESS']])
        // asserts
        assertEquals('Message not set correctly', 'SUCCESS: Job p <http://build.url|#1>', slackCallMap.message.toString())
        assertNull('Channel not set correctly', slackCallMap.channel)
        assertEquals('Color not set correctly', '#8cc04f', slackCallMap.color)
        assertJobStatusSuccess()
    }

    @Test
    void testNotificationBuildSuccessCustomChannel() throws Exception {
        stepRule.step.slackSendNotification(script: [currentBuild: [result: 'SUCCCESS']], channel: 'Test')
        // asserts
        assertEquals('Channel not set correctly', 'Test', slackCallMap.channel)
        assertJobStatusSuccess()
    }

    @Test
    void testNotificationBuildFailed() throws Exception {
        stepRule.step.slackSendNotification(script: [currentBuild: [result: 'FAILURE']])
        // asserts
        assertEquals('Message not set correctly', 'FAILURE: Job p <http://build.url|#1>', slackCallMap.message.toString())
        assertEquals('Color not set correctly', '#d54c53', slackCallMap.color)
    }

    @Test
    void testNotificationBuildStatusNull() throws Exception {
        stepRule.step.slackSendNotification(script: [currentBuild: [:]])
        // asserts
        assertTrue('Missing build status not detected', loggingRule.log.contains('currentBuild.result is not set. Skipping Slack notification'))
        assertJobStatusSuccess()
    }

    @Test
    void testNotificationCustomMessageAndColor() throws Exception {
        stepRule.step.slackSendNotification(script: [currentBuild: [:]], message: 'Custom Message', color: '#AAAAAA')
        // asserts
        assertEquals('Custom message not set correctly', 'Custom Message', slackCallMap.message.toString())
        assertEquals('Custom color not set correctly', '#AAAAAA', slackCallMap.color)
        assertJobStatusSuccess()
    }

    @Test
    void testNotificationWithCustomCredentials() throws Exception {
        stepRule.step.slackSendNotification(
            script: [currentBuild: [:]],
            message: 'I am no Message',
            baseUrl: 'https://my.base.url',
            credentialsId: 'MY_TOKEN_ID'
        )
        // asserts
        assertEquals('Custom base url not set correctly', 'https://my.base.url', slackCallMap.baseUrl)
        assertEquals('Custom token id not set correctly', 'MY_TOKEN_ID', slackCallMap.tokenCredentialId)
        assertJobStatusSuccess()
    }
}
