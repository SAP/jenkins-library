import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString

import org.hamcrest.Matchers
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.Utils
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
        Utils.metaClass.echo = { def m -> }
        helper.registerAllowedMethod('addBadge', [Map], {return})
        helper.registerAllowedMethod('createSummary', [Map], {return})
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
    }

    @Test
    public void changeDocumentIdNotProvidedSOLMANTest() {

        ChangeManagement cm = new ChangeManagement(nullScript) {
            String getChangeDocumentId(String from,
                                       String to,
                                       String label,
                                       String format) {
                                throw new ChangeManagementException('Cannot retrieve change documentId')
            }
        }

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("Change document id not provided (parameter: 'changeDocumentId' provided to the step call or via commit history).")

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
        thrown.expectMessage("Transport request id not provided (parameter: 'transportRequestId' provided to the step call or via commit history).")

        stepRule.step.transportRequestRelease(script: nullScript, changeDocumentId: '001', cmUtils: cm)
    }

    @Test
    public void releaseTransportRequestFailsSOLMANTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Something went wrong")

        ChangeManagement cm = new ChangeManagement(nullScript) {

            void releaseTransportRequestSOLMAN(
                                         Map docker,
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
                                         Map docker,
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
                                rfc: [
                                    dockerImage: 'rfc',
                                    dockerOptions: [],
                                ],
                            ]

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void releaseTransportRequestRFC(
                Map docker,
                String transportRequestId,
                String endpoint,
                String developmentInstance,
                String developmentClient,
                String credentialsId,
                boolean verbose) {

                receivedParameters = [
                    docker: docker,
                    transportRequestId: transportRequestId,
                    endpoint: endpoint,
                    developmentInstance: developmentInstance,
                    developmentClient: developmentClient,
                    credentialsId: credentialsId,
                    verbose: verbose,
                ]
            }
        }

        stepRule.step.transportRequestRelease(
            script: nullScript,
            transportRequestId: '002',
            changeManagement: [
                rfc: [
                    developmentClient: '003',
                    developmentInstance: '002',
                ]
            ],
            verbose: true,
            cmUtils: cm)

        assert receivedParameters == [
                    docker: [
                        image: 'rfc',
                        options: [],
                        envVars: [:],
                        pullImage: true,
                    ],
                    transportRequestId: '002',
                    endpoint: 'https://example.org/rfc',
                    developmentInstance: '002',
                    developmentClient: '003',
                    credentialsId: 'CM',
                    'verbose': true,
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
                Map docker,
                String transportRequestId,
                String endpoint,
                String credentialsId,
                String clientOpts = '') {

                receivedParameters = [
                    docker: docker,
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
                    docker: [
                        image:'ppiper/cm-client:2.0.1.0',
                        options:[],
                        envVars:[:],
                        pullImage:true,
                    ],
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
                Map docker,
                String transportRequestId,
                String endpoint,
                String developmentInstance,
                String developmentClient,
                String credentialsId,
                boolean verbose) {

                throw new ChangeManagementException('Failed releasing transport request.')
            }
        }

        stepRule.step.transportRequestRelease(
            script: nullScript,
            transportRequestId: '002',
            changeManagement: [
                rfc: [
                    developmentClient: '003',
                    developmentInstance: '002'
                ]
            ],
            cmUtils: cm)

    }

    @Test
    public void releaseTransportRequestSanityChecksSOLMANTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage(allOf(
            containsString('ERROR - NO VALUE AVAILABLE FOR'),
            containsString('changeManagement/endpoint')))

        // changeDocumentId and transportRequestId are not checked
        // by the sanity checks here since they are looked up from
        // commit history in case they are not provided.

        nullScript
            .commonPipelineEnvironment
                .configuration = null

        stepRule.step.transportRequestRelease(
            script: nullScript,
            changeManagement: [type: 'SOLMAN']
        )
    }

    @Test
    public void releaseTransportRequestSanityChecksCTSTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage(allOf(
            containsString('ERROR - NO VALUE AVAILABLE FOR'),
            containsString('changeManagement/endpoint')))

        nullScript
            .commonPipelineEnvironment
                .configuration = null

        stepRule.step.transportRequestRelease(
            script: nullScript,
            changeManagement: [type: 'CTS']
        )
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
                                         Map docker,
                                         String changeId,
                                         String transportRequestId,
                                         String endpoint,
                                         String credentialsId,
                                         String clientOpts) {

                receivedParams.docker = docker
                receivedParams.changeId = changeId
                receivedParams.transportRequestId = transportRequestId
                receivedParams.endpoint = endpoint
                receivedParams.credentialsId = credentialsId
                receivedParams.clientOpts = clientOpts
            }
        }

        stepRule.step.transportRequestRelease(script: nullScript, changeDocumentId: '001', transportRequestId: '002', cmUtils: cm)

        assert receivedParams == [
                                  docker: [
                                      image: 'ppiper/cm-client:2.0.1.0',
                                      pullImage: true,
                                      envVars: [:],
                                      options: [],
                                  ],
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
