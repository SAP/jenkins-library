import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString

import java.util.Map

import org.hamcrest.Matchers
import org.hamcrest.core.StringContains
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

public class TransportRequestCreateTest extends BasePiperTest {

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
                                          endpoint: 'https://example.org/cm',
                                          clientOpts: '-DmyProp=myVal',
                                          changeDocumentLabel: 'ChangeId\\s?:',
                                          git: [from: 'origin/master',
                                                to: 'HEAD',
                                                format: '%b']
                                         ]
                                     ]
                                 ]
    }

    @Test
    public void changeIdNotProvidedSOLANTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("Change document id not provided (parameter: 'changeDocumentId' or via commit history).")
        ChangeManagement cm = new ChangeManagement(nullScript) {
            String getChangeDocumentId(
                                       String from,
                                       String to,
                                       String label,
                                       String format
                                      ) {
                                          throw new ChangeManagementException('Cannot retrieve changeId from git commits.')
                                      }
        }

        stepRule.step.transportRequestCreate(script: nullScript, developmentSystemId: '001', cmUtils: cm)
    }

    @Test
    public void developmentSystemIdNotProvidedSOLMANTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("ERROR - NO VALUE AVAILABLE FOR developmentSystemId")

        stepRule.step.transportRequestCreate(script: nullScript, changeDocumentId: '001')
    }

    @Test
    public void createTransportRequestFailureSOLMANTest() {

        ChangeManagement cm = new ChangeManagement(nullScript) {

            String createTransportRequestSOLMAN(
                                          String changeId,
                                          String developmentSystemId,
                                          String cmEndpoint,
                                          String credentialId,
                                          String clientOpts) {

                    throw new ChangeManagementException('Exception message.')
            }
        }


        thrown.expect(AbortException)
        thrown.expectMessage("Exception message.")

        stepRule.step.transportRequestCreate(script: nullScript, changeDocumentId: '001', developmentSystemId: '001', cmUtils: cm)
    }

    @Test
    public void createTransportRequestSuccessSOLMANTest() {

        def result = [:]

        ChangeManagement cm = new ChangeManagement(nullScript) {

            String createTransportRequestSOLMAN(
                                          String changeId,
                                          String developmentSystemId,
                                          String cmEndpoint,
                                          String credentialId,
                                          String clientOpts) {

                result.changeId = changeId
                result.developmentSystemId = developmentSystemId
                result.cmEndpoint = cmEndpoint
                result.credentialId = credentialId
                result.clientOpts = clientOpts
                return '001'
            }
        }

        stepRule.step.transportRequestCreate(script: nullScript, changeDocumentId: '001', developmentSystemId: '001', cmUtils: cm)

        assert nullScript.commonPipelineEnvironment.getTransportRequestId() == '001'
        assert result == [changeId: '001',
                         developmentSystemId: '001',
                         cmEndpoint: 'https://example.org/cm',
                         credentialId: 'CM',
                         clientOpts: '-DmyProp=myVal'
                         ]

        assert loggingRule.log.contains("[INFO] Creating transport request for change document '001' and development system '001'.")
        assert loggingRule.log.contains("[INFO] Transport Request '001' has been successfully created.")
    }

    @Test
    public void createTransportRequestSuccessCTSTest() {

        def result = [:]

        ChangeManagement cm = new ChangeManagement(nullScript) {

            String createTransportRequestCTS(
                String transportType,
                String targetSystemId,
                String description,
                String endpoint,
                String credentialsId,
                String clientOpts) {
                result.transportType = transportType
                result.targetSystemId = targetSystemId
                result.description = description
                result.endpoint = endpoint
                result.credentialsId = credentialsId
                result.clientOpts = clientOpts
                return '001'
            }
        }

        stepRule.step.call(script: nullScript,
                        transportType: 'W',
                        targetSystem: 'XYZ',
                        description: 'desc',
                        changeManagement: [type: 'CTS'],
                        cmUtils: cm)

        assert nullScript.commonPipelineEnvironment.getTransportRequestId() == '001'
        assert result == [transportType: 'W',
                         targetSystemId: 'XYZ',
                         description: 'desc',
                         endpoint: 'https://example.org/cm',
                         credentialsId: 'CM',
                         clientOpts: '-DmyProp=myVal'
                         ]

        assert loggingRule.log.contains("[INFO] Creating transport request.")
        assert loggingRule.log.contains("[INFO] Transport Request '001' has been successfully created.")
    }

    @Test
    public void createTransportRequestSuccessRFCTest() {

        def result = [:]

        ChangeManagement cm = new ChangeManagement(nullScript) {

            String createTransportRequestRFC(
                Map docker,
                String endpoint,
                String developmentClient,
                String developmentInstance,
                String credentialsId,
                String description) {

                result.docker = docker
                result.endpoint = endpoint
                result.developmentClient = developmentClient
                result.developmentInstance= developmentInstance
                result.credentialsId = credentialsId
                result.description = description

                return '001'
            }
        }

        stepRule.step.transportRequestCreate(
            script: nullScript,
            changeManagement: [
                type: 'RFC',
                rfc: [
                    developmentInstance: '01',
                    developmentClient: '001',
                ],
                endpoint: 'https://example.org/rfc',
            ],
            developmentSystemId: '001',
            description: '',
            cmUtils: cm)

        assert nullScript.commonPipelineEnvironment.getTransportRequestId() == '001'
        assert result == [
            docker: [
                image: 'rfc',
                options: [],
                envVars: [:],
                pullImage: true
            ],
            endpoint: 'https://example.org/rfc',
            developmentClient: '01',
            developmentInstance: '001',
            credentialsId: 'CM',
            description: ''
        ]

        assert loggingRule.log.contains("[INFO] Creating transport request.")
        assert loggingRule.log.contains("[INFO] Transport Request '001' has been successfully created.")
    }

    @Test
    public void createTransportRequestFailureRFCTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('upload failed')

        ChangeManagement cm = new ChangeManagement(nullScript) {

            String createTransportRequestRFC(
                Map docker,
                String endpoint,
                String developmentClient,
                String developmentInstance,
                String credentialsId,
                String description) {

                throw new ChangeManagementException('upload failed')
            }
        }

        stepRule.step.transportRequestCreate(
            script: nullScript,
            changeManagement: [
                type: 'RFC',
                rfc: [
                    developmentInstance: '01',
                    developmentClient: '001',
                ],
                endpoint: 'https://example.org/rfc',
            ],
            developmentSystemId: '001',
            description: '',
            cmUtils: cm)
    }

    @Test
    public void createTransportRequestSanityChecksRFCTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage(allOf(
            containsString('changeManagement/rfc/developmentInstance'),
            containsString('changeManagement/rfc/developmentClient'),
            ))
        stepRule.step.transportRequestCreate(
            script: nullScript,
            changeManagement: [
                type: 'RFC',
            ])
    }

    @Test
    public void cmIntegrationSwichtedOffTest() {

        loggingRule.expect('[INFO] Change management integration intentionally switched off.')

        stepRule.step.transportRequestCreate(script: nullScript,
            changeManagement: [type: 'NONE'])
    }
}
