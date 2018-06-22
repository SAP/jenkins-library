import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.GitUtils
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException
import util.BasePiperTest
import util.JenkinsStepRule
import util.Rules

class CheckChangeInDevelopmentTest extends BasePiperTest {

    private ExpectedException thrown = ExpectedException.none()
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jsr)

    @Before
    public void setup() {
        helper.registerAllowedMethod('usernamePassword', [Map], { Map m ->
            binding.setProperty('username', 'defaultUser')
            binding.setProperty('password', '********')
        })

        helper.registerAllowedMethod('withCredentials', [List, Closure], { List l, Closure c ->
            c()
        })
    }

    @After
    public void tearDown() {
        cmUtilReceivedParams.clear()
    }

    private Map cmUtilReceivedParams = [:]

    @Test
    public void changeIsInStatusDevelopmentTest() {

        ChangeManagement cm = getChangeManagementUtils(true)
        boolean inDevelopment = jsr.step.checkChangeInDevelopment(
                                    cmUtils: cm,
                                    endpoint: 'https://example.org/cm')

        assert inDevelopment

        assert cmUtilReceivedParams == [
            changeId: '001',
            endpoint: 'https://example.org/cm',
            userName: 'defaultUser',
            password: '********'
        ]
    }

    @Test
    public void changeIsNotInStatusDevelopmentTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Change '001' is not in status 'in development'")

        ChangeManagement cm = getChangeManagementUtils(false)
        jsr.step.checkChangeInDevelopment(
            cmUtils: cm,
            endpoint: 'https://example.org/cm')
    }

    @Test
    public void changeIsNotInStatusDevelopmentButWeWouldLikeToSkipFailureTest() {

        ChangeManagement cm = getChangeManagementUtils(false)
        boolean inDevelopment = jsr.step.checkChangeInDevelopment(
                                    cmUtils: cm,
                                    endpoint: 'https://example.org/cm',
                                    failIfStatusIsNotInDevelopment: false)
        assert !inDevelopment
    }

    @Test
    public void changeDocumentIdRetrievalFailsTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('Something went wrong')

        ChangeManagement cm = new ChangeManagement(nullScript, null) {

            String getChangeDocumentId(
                String filter,
                String from,
                String to,
                String format) {
                throw new ChangeManagementException('Something went wrong')
            }
        }

        jsr.step.checkChangeInDevelopment(
            cmUtils: cm,
            endpoint: 'https://example.org/cm')
    }

    @Test
    public void nullChangeDocumentIdTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("ChangeId is null or empty.")

        ChangeManagement cm = getChangeManagementUtils(false, null)
        jsr.step.checkChangeInDevelopment(
            cmUtils: cm,
            endpoint: 'https://example.org/cm')
    }

    @Test
    public void emptyChangeDocumentIdTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("ChangeId is null or empty.")

        ChangeManagement cm = getChangeManagementUtils(false, '')
        jsr.step.checkChangeInDevelopment(
            cmUtils: cm,
            endpoint: 'https://example.org/cm')
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

            boolean isChangeInDevelopment(String changeId, String endpoint, String userName, String password) {
                cmUtilReceivedParams.changeId = changeId
                cmUtilReceivedParams.endpoint = endpoint
                cmUtilReceivedParams.userName = userName
                cmUtilReceivedParams.password = password

                return inDevelopment
            }
        }
    }
}
