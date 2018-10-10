#!groovy
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

    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jlr)
        .around(jsr)

    @Before
    void init() throws Exception {
        helper.registerAllowedMethod("slackSend", [Map.class], {m -> slackCallMap = m})
    }

    @Test
    void testNotificationBuildSuccessDefaultChannel() throws Exception {
        jsr.step.slackSendNotification(script: [currentBuild: [result: 'SUCCESS']])
        // asserts
        assertEquals('Message not set correctly', 'SUCCESS: Job p <http://build.url|#1>', slackCallMap.message.toString())
        assertNull('Channel not set correctly', slackCallMap.channel)
        assertEquals('Color not set correctly', '#008000', slackCallMap.color)
        assertJobStatusSuccess()
    }

    @Test
    void testNotificationBuildSuccessCustomChannel() throws Exception {
        jsr.step.slackSendNotification(script: [currentBuild: [result: 'SUCCCESS']], channel: 'Test')
        // asserts
        assertEquals('Channel not set correctly', 'Test', slackCallMap.channel)
        assertJobStatusSuccess()
    }

    @Test
    void testNotificationBuildFailed() throws Exception {
        jsr.step.slackSendNotification(script: [currentBuild: [result: 'FAILURE']])
        // asserts
        assertEquals('Message not set correctly', 'FAILURE: Job p <http://build.url|#1>', slackCallMap.message.toString())
        assertEquals('Color not set correctly', '#E60000', slackCallMap.color)
    }

    @Test
    void testNotificationBuildStatusNull() throws Exception {
        jsr.step.slackSendNotification(script: [currentBuild: [:]])
        // asserts
        assertTrue('Missing build status not detected', jlr.log.contains('currentBuild.result is not set. Skipping Slack notification'))
        assertJobStatusSuccess()
    }

    @Test
    void testNotificationCustomMessageAndColor() throws Exception {
        jsr.step.slackSendNotification(script: [currentBuild: [:]], message: 'Custom Message', color: '#AAAAAA')
        // asserts
        assertEquals('Custom message not set correctly', 'Custom Message', slackCallMap.message.toString())
        assertEquals('Custom color not set correctly', '#AAAAAA', slackCallMap.color)
        assertJobStatusSuccess()
    }

    @Test
    void testNotificationWithCustomCredentials() throws Exception {
        jsr.step.slackSendNotification(
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
