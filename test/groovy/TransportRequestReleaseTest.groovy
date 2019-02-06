import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString

import org.hamcrest.Matchers
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
import util.JenkinsDockerExecuteRule
import util.JenkinsStepRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.Rules

import hudson.AbortException
import hudson.scm.NullSCM

public class TransportRequestReleaseTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(stepRule)
        .around(loggingRule)
        .around(new JenkinsCredentialsRule(this)
            .withCredentials('CM', 'anonymous', '********'))

    @Before
    public void setup() {

        nullScript.commonPipelineEnvironment.configuration = [general:
                                     [changeManagement:
                                         [
                                          credentialsId: 'CM',
                                          type: 'SOLMAN',
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

        stepRule.step.transportRequestRelease(script: nullScript, transportRequestId: '001', cmUtils: cm)
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

        stepRule.step.transportRequestRelease(script: nullScript, changeDocumentId: '001', cmUtils: cm)
    }

    @Test
    public void releaseTransportRequestFailsSOLMANTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Something went wrong")

        ChangeManagement cm = new ChangeManagement(nullScript) {

            void releaseTransportRequestSOLMAN(
                                         String changeId,
                                         String transportRequestId,
                                         String endpoint,
                                         String credentialsId,
                                         String clientOpts) {

                throw new ChangeManagementException('Something went wrong')
            }
        }

        stepRule.step.transportRequestRelease(script: nullScript, changeDocumentId: '001', transportRequestId: '001', cmUtils: cm)
    }

    @Test
    public void releaseTransportRequestFailsCTSTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Something went wrong")

        nullScript
            .commonPipelineEnvironment
                .configuration
                    .general
                        .changeManagement
                            .type = 'CTS'

        ChangeManagement cm = new ChangeManagement(nullScript) {

            void releaseTransportRequestCTS(
                                         String transportRequestId,
                                         String endpoint,
                                         String credentialsId,
                                         String clientOpts) {

                throw new ChangeManagementException('Something went wrong')
            }
        }

        stepRule.step.transportRequestRelease(
            script: nullScript,
            transportRequestId: '001',
            cmUtils: cm)
    }

    @Test
    public void releaseTransportRequestSuccessRFCTest() {

        def receivedParameters

        nullScript
            .commonPipelineEnvironment
                .configuration
                    .general
                        .changeManagement =
                            [
                                credentialsId: 'CM',
                                type: 'RFC',
                                endpoint: 'https://example.org/rfc',
                                rfc: [dockerImage: 'rfc']
                            ]

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void releaseTransportRequestRFC(
                String dockerImage,
                List dockerOptions,
                String transportRequestId,
                String endpoint,
                String developmentClient,
                String credentialsId) {

                receivedParameters = [
                    dockerImage: dockerImage,
                    dockerOptions: dockerOptions,
                    transportRequestId: transportRequestId,
                    endpoint: endpoint,
                    developmentClient: developmentClient,
                    credentialsId: credentialsId,
                ]
            }
        }

        stepRule.step.transportRequestRelease(
            script: nullScript,
            transportRequestId: '002',
            developmentClient: '003',
            cmUtils: cm)

        assert receivedParameters == [
                    dockerImage: 'rfc',
                    dockerOptions: [],
                    transportRequestId: '002',
                    endpoint: 'https://example.org/rfc',
                    developmentClient: '003',
                    credentialsId: 'CM',
                ]
    }

    @Test
    public void releaseTransportRequestSuccessCTSTest() {

        def receivedParameters

        nullScript
            .commonPipelineEnvironment
                .configuration
                    .general
                        .changeManagement =
                            [
                                credentialsId: 'CM',
                                type: 'CTS',
                                endpoint: 'https://example.org/cts'
                            ]

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void releaseTransportRequestCTS(
                String transportRequestId,
                String endpoint,
                String credentialsId,
                String clientOpts = '') {

                receivedParameters = [
                    transportRequestId: transportRequestId,
                    endpoint: endpoint,
                    credentialsId: credentialsId,
                    clientOpts: clientOpts
                ]
            }
        }

        stepRule.step.transportRequestRelease(
            script: nullScript,
            transportRequestId: '002',
            cmUtils: cm)

        assert receivedParameters == [
                    transportRequestId: '002',
                    endpoint: 'https://example.org/cts',
                    credentialsId: 'CM',
                    clientOpts: ''
                ]
    }

    @Test
    public void releaseTransportRequestFailsRFCTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('Failed releasing transport request.')

        nullScript
            .commonPipelineEnvironment
                .configuration
                    .general
                        .changeManagement =
                            [
                                credentialsId: 'CM',
                                type: 'RFC',
                                endpoint: 'https://example.org/rfc',
                                rfc: [dockerImage: 'rfc']
                            ]

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void releaseTransportRequestRFC(
                String dockerImage,
                List dockerOptions,
                String transportRequestId,
                String endpoint,
                String developmentClient,
                String credentialsId) {

                throw new ChangeManagementException('Failed releasing transport request.')
            }
        }

        stepRule.step.transportRequestRelease(
            script: nullScript,
            transportRequestId: '002',
            developmentClient: '003',
            cmUtils: cm)

    }

    @Test
    public void releaseTransportRequestSanityChecksRFCTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage(allOf(
            containsString('ERROR - NO VALUE AVAILABLE FOR:'),
            containsString('changeManagement/endpoint'),
            containsString('developmentClient')))

        nullScript
            .commonPipelineEnvironment
                .configuration = null

        stepRule.step.transportRequestRelease(
            script: nullScript,
            changeManagement: [type: 'RFC'],
            transportRequestId: '002')
    }

    @Test
    public void releaseTransportRequestSuccessSOLMANTest() {

        // Here we test only the case where the transportRequestId is
        // provided via parameters. The other cases are tested by
        // corresponding tests for StepHelpers#getTransportRequestId(./.)

        loggingRule.expect("[INFO] Closing transport request '002' for change document '001'.")
        loggingRule.expect("[INFO] Transport Request '002' has been successfully closed.")

        Map receivedParams = [:]

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void releaseTransportRequestSOLMAN(
                                         String changeId,
                                         String transportRequestId,
                                         String endpoint,
                                         String credentialsId,
                                         String clientOpts) {

                receivedParams.changeId = changeId
                receivedParams.transportRequestId = transportRequestId
                receivedParams.endpoint = endpoint
                receivedParams.credentialsId = credentialsId
                receivedParams.clientOpts = clientOpts
            }
        }

        stepRule.step.transportRequestRelease(script: nullScript, changeDocumentId: '001', transportRequestId: '002', cmUtils: cm)

        assert receivedParams == [
                                  changeId: '001',
                                  transportRequestId: '002',
                                  endpoint: 'https://example.org/cm',
                                  credentialsId: 'CM',
                                  clientOpts: '']
    }

    @Test
    public void cmIntegrationSwichtedOffTest() {

        loggingRule.expect('[INFO] Change management integration intentionally switched off.')

        stepRule.step.transportRequestRelease(script: nullScript,
            changeManagement: [type: 'NONE'])
    }
}
