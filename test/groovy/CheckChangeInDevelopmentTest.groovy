import org.junit.After
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.GitUtils
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException
import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

class CheckChangeInDevelopmentTest extends BasePiperTest {

    private ExpectedException thrown = ExpectedException.none()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(stepRule)
        .around(loggingRule)
        .around(new JenkinsCredentialsRule(this)
        .withCredentials('CM', 'anonymous', '********'))

    @After
    public void tearDown() {
        cmUtilReceivedParams.clear()
    }

    private Map cmUtilReceivedParams = [:]

    @Test
    public void changeIsInStatusDevelopmentTest() {

        ChangeManagement cm = getChangeManagementUtils(true)
        stepRule.step.checkChangeInDevelopment(
            script: nullScript,
            cmUtils: cm,
            changeManagement: [
                type: 'SOLMAN',
                endpoint: 'https://example.org/cm'],
            failIfStatusIsNotInDevelopment: true)

        assert cmUtilReceivedParams == [
            changeId: '001',
            endpoint: 'https://example.org/cm',
            credentialsId: 'CM',
            cmclientOpts: ''
        ]

        // no exception in thrown, so the change is in status 'in development'.
    }

    @Test
    public void changeIsNotInStatusDevelopmentTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Change '001' is not in status 'in development'")

        ChangeManagement cm = getChangeManagementUtils(false)
        stepRule.step.checkChangeInDevelopment(
            script: nullScript,
            cmUtils: cm,
            changeManagement: [type: 'SOLMAN',
                               endpoint: 'https://example.org/cm'])
    }

    @Test
    public void changeIsNotInStatusDevelopmentButWeWouldLikeToSkipFailureTest() {

        ChangeManagement cm = getChangeManagementUtils(false)
        boolean inDevelopment = stepRule.step.checkChangeInDevelopment(
                                    script: nullScript,
                                    cmUtils: cm,
                                    changeManagement: [endpoint: 'https://example.org/cm'],
                                    failIfStatusIsNotInDevelopment: false)
        assert !inDevelopment
    }

    @Test
    public void changeDocumentIdRetrievalFailsTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("No changeDocumentId provided. Neither via parameter 'changeDocumentId' nor via " +
                             "label 'ChangeDocument\\s?:' in commit range [from: origin/master, to: HEAD].")

        ChangeManagement cm = new ChangeManagement(nullScript, null) {

            String getChangeDocumentId(
                String filter,
                String from,
                String to,
                String format) {
                throw new ChangeManagementException('Something went wrong')
            }
        }

        stepRule.step.checkChangeInDevelopment(
            script: nullScript,
            cmUtils: cm,
            changeManagement: [type: 'SOLMAN',
                               endpoint: 'https://example.org/cm'])
    }

    @Test
    public void nullChangeDocumentIdTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("No changeDocumentId provided. Neither via parameter 'changeDocumentId' " +
                             "nor via label 'ChangeDocument\\s?:' in commit range " +
                             "[from: origin/master, to: HEAD].")

        ChangeManagement cm = getChangeManagementUtils(false, null)
        stepRule.step.checkChangeInDevelopment(
            script: nullScript,
            cmUtils: cm,
            changeManagement: [endpoint: 'https://example.org/cm',
                               type: 'SOLMAN'])
    }

    @Test
    public void emptyChangeDocumentIdTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("No changeDocumentId provided. Neither via parameter 'changeDocumentId' " +
                             "nor via label 'ChangeDocument\\s?:' in commit range " +
                             "[from: origin/master, to: HEAD].")

        ChangeManagement cm = getChangeManagementUtils(false, '')
        stepRule.step.checkChangeInDevelopment(
            script: nullScript,
            cmUtils: cm,
            changeManagement: [type: 'SOLMAN',
                               endpoint: 'https://example.org/cm'])
    }

    @Test
    public void cmIntegrationSwichtedOffTest() {

        loggingRule.expect('[INFO] Change management integration intentionally switched off.')

        stepRule.step.checkChangeInDevelopment(
            script: nullScript,
            changeManagement: [type: 'NONE'])

    }

    private ChangeManagement getChangeManagementUtils(boolean inDevelopment, String changeDocumentId = '001') {

        return new ChangeManagement(nullScript, null) {

            String getChangeDocumentId(
                String filter,
                String from,
                String to,
                String format) {
                return changeDocumentId
            }

            boolean isChangeInDevelopment(String changeId, String endpoint, String credentialsId, String cmclientOpts) {
                cmUtilReceivedParams.changeId = changeId
                cmUtilReceivedParams.endpoint = endpoint
                cmUtilReceivedParams.credentialsId = credentialsId
                cmUtilReceivedParams.cmclientOpts = cmclientOpts

                return inDevelopment
            }
        }
    }
}
