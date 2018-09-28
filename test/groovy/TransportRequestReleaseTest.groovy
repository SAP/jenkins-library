import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.cm.BackendType
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsStepRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.Rules

import hudson.AbortException
import hudson.scm.NullSCM

public class TransportRequestReleaseTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(jsr)
        .around(jlr)
        .around(new JenkinsCredentialsRule(this)
            .withCredentials('CM', 'anonymous', '********'))

    @Before
    public void setup() {

        nullScript.commonPipelineEnvironment.configuration = [general:
                                     [changeManagement:
                                         [
                                          credentialsId: 'CM',
                                          endpoint: 'https://example.org/cm'
                                         ]
                                     ]
                                 ]
    }

    @Test
    public void changeIdNotProvidedTest() {

        ChangeManagement cm = new ChangeManagement(nullScript) {
            String getChangeDocumentId(String from,
                                       String to,
                                       String label,
                                       String format) {
                                throw new ChangeManagementException('Cannot retrieve change documentId')
            }
        }

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("Change document id not provided (parameter: 'changeDocumentId' or via commit history).")

        jsr.step.call(script: nullScript, transportRequestId: '001', cmUtils: cm)
    }

    @Test
    public void transportRequestIdNotProvidedTest() {

        ChangeManagement cm = new ChangeManagement(nullScript) {
            String getTransportRequestId(String from,
                                         String to,
                                         String label,
                                         String format) {
                throw new ChangeManagementException('Cannot retrieve transportRequestId')
            }
        }

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("Transport request id not provided (parameter: 'transportRequestId' or via commit history).")

        jsr.step.call(script: nullScript, changeDocumentId: '001', cmUtils: cm)
    }

    @Test
    public void releaseTransportRequestFailureTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Something went wrong")

        ChangeManagement cm = new ChangeManagement(nullScript) {

            void releaseTransportRequest(BackendType type,
                                         String changeId,
                                         String transportRequestId,
                                         String endpoint,
                                         String credentialsId,
                                         String clientOpts) {

                throw new ChangeManagementException('Something went wrong')
            }
        }

        jsr.step.call(script: nullScript, changeDocumentId: '001', transportRequestId: '001', cmUtils: cm)
    }

    @Test
    public void releaseTransportRequestSuccessTest() {

        jlr.expect("[INFO] Closing transport request '002' for change document '001'.")
        jlr.expect("[INFO] Transport Request '002' has been successfully closed.")

        Map receivedParams = [:]

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void releaseTransportRequest(BackendType type,
                                         String changeId,
                                         String transportRequestId,
                                         String endpoint,
                                         String credentialsId,
                                         String clientOpts) {

                receivedParams.type = type
                receivedParams.changeId = changeId
                receivedParams.transportRequestId = transportRequestId
                receivedParams.endpoint = endpoint
                receivedParams.credentialsId = credentialsId
                receivedParams.clientOpts = clientOpts
            }
        }

        jsr.step.call(script: nullScript, changeDocumentId: '001', transportRequestId: '002', cmUtils: cm)

        assert receivedParams == [type: BackendType.SOLMAN,
                                  changeId: '001',
                                  transportRequestId: '002',
                                  endpoint: 'https://example.org/cm',
                                  credentialsId: 'CM',
                                  clientOpts: '']
    }

    @Test
    public void cmIntegrationSwichtedOffTest() {

        jlr.expect('[INFO] Change management integration intentionally switched off.')

        jsr.step.call(
            changeManagement: [type: 'NONE'])
    }
}
