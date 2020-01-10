import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

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
    }

    @Test
    void testNotificationBuildSuccessDefaultChannel() throws Exception {
        nullScript.currentBuild = [result: 'SUCCESS']
        stepRule.step.slackSendNotification(script: nullScript)
        // asserts
        assertEquals('Message not set correctly', 'SUCCESS: Job p <http://build.url|#1>', slackCallMap.message.toString())
        assertNull('Channel not set correctly', slackCallMap.channel)
        assertEquals('Color not set correctly', '#008000', slackCallMap.color)
        assertJobStatusSuccess()
    }

    @Test
    void testNotificationBuildSuccessCustomChannel() throws Exception {
        nullScript.currentBuild = [result: 'SUCCCESS']
        stepRule.step.slackSendNotification(script: nullScript, channel: 'Test')
        // asserts
        assertEquals('Channel not set correctly', 'Test', slackCallMap.channel)
        assertJobStatusSuccess()
    }

    @Test
    void testNotificationBuildFailed() throws Exception {
        nullScript.currentBuild = [result: 'FAILURE']
        stepRule.step.slackSendNotification(script: nullScript)
        // asserts
        assertEquals('Message not set correctly', 'FAILURE: Job p <http://build.url|#1>', slackCallMap.message.toString())
        assertEquals('Color not set correctly', '#E60000', slackCallMap.color)
    }

    @Test
    void testNotificationBuildStatusNull() throws Exception {
        nullScript.currentBuild = [:]
        stepRule.step.slackSendNotification(script: nullScript)
        // asserts
        assertTrue('Missing build status not detected', loggingRule.log.contains('currentBuild.result is not set. Skipping Slack notification'))
        assertJobStatusSuccess()
    }

    @Test
    void testNotificationCustomMessageAndColor() throws Exception {
        nullScript.currentBuild = [:]
        stepRule.step.slackSendNotification(script: nullScript, message: 'Custom Message', color: '#AAAAAA')
        // asserts
        assertEquals('Custom message not set correctly', 'Custom Message', slackCallMap.message.toString())
        assertEquals('Custom color not set correctly', '#AAAAAA', slackCallMap.color)
        assertJobStatusSuccess()
    }

    @Test
    void testNotificationWithCustomCredentials() throws Exception {
        nullScript.currentBuild = [:]
        stepRule.step.slackSendNotification(
            script: nullScript,
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
