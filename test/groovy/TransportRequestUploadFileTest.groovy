import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString

import java.util.List
import java.util.Map

import org.hamcrest.Matchers
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.JenkinsUtils
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

public class TransportRequestUploadFileTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(thrown)
        .around(new JenkinsReadYamlRule(this))
        .around(stepRule)
        .around(loggingRule)
        .around(new JenkinsCredentialsRule(this)
            .withCredentials('CM', 'anonymous', '********'))

    private Map cmUtilReceivedParams = [:]

    @Before
    public void setup() {

        cmUtilReceivedParams.clear()

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
    public void changeDocumentIdNotProvidedSOLMANTest() {

        // we expect the failure only for SOLMAN (which is the default).
        // Use case for CTS without change document id is checked by the
        // straight forward test case for CTS

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("Change document id not provided (parameter: 'changeDocumentId' or via commit history).")

        ChangeManagement cm = new ChangeManagement(nullScript) {
            String getChangeDocumentId(
                                       String from,
                                       String to,
                                       String pattern,
                                       String format
                                    ) {
                                        throw new ChangeManagementException('Cannot retrieve changeId from git commits.')
                                      }
        }

        stepRule.step.transportRequestUploadFile(script: nullScript, transportRequestId: '001', applicationId: 'app', filePath: '/path', cmUtils: cm)
    }

    @Test
    public void transportRequestIdNotProvidedTest() {

        ChangeManagement cm = new ChangeManagement(nullScript) {
            String getTransportRequestId(
                                       String from,
                                       String to,
                                       String pattern,
                                       String format
                                    ) {
                                        throw new ChangeManagementException('Cannot retrieve transport request id from git commits.')
                                    }
        }

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("Transport request id not provided (parameter: 'transportRequestId' or via commit history).")

        stepRule.step.transportRequestUploadFile(script: nullScript, changeDocumentId: '001', applicationId: 'app', filePath: '/path', cmUtils: cm)
    }

    @Test
    public void applicationIdNotProvidedSOLMANTest() {

        // we expect the failure only for SOLMAN (which is the default).
        // Use case for CTS without applicationId is checked by the
        // straight forward test case for CTS

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("ERROR - NO VALUE AVAILABLE FOR applicationId")

        stepRule.step.transportRequestUploadFile(script: nullScript, changeDocumentId: '001', transportRequestId: '001', filePath: '/path')
    }

    @Test
    public void filePathNotProvidedTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("ERROR - NO VALUE AVAILABLE FOR filePath")

        stepRule.step.transportRequestUploadFile(script: nullScript, changeDocumentId: '001', transportRequestId: '001', applicationId: 'app')
    }

    @Test
    public void uploadFileToTransportRequestSOLMANFailureTest() {

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void uploadFileToTransportRequestSOLMAN(
                                              String changeId,
                                              String transportRequestId,
                                              String applicationId,
                                              String filePath,
                                              String endpoint,
                                              String credentialsId,
                                              String cmclientOpts) {
                throw new ChangeManagementException('Exception message')
            }
        }

        thrown.expect(AbortException)
        thrown.expectMessage("Exception message")

        stepRule.step.transportRequestUploadFile(script: nullScript,
                      changeDocumentId: '001',
                      transportRequestId: '001',
                      applicationId: 'app',
                      filePath: '/path',
                      cmUtils: cm)
    }

    @Test
    public void uploadFileToTransportRequestCTSSuccessTest() {

        loggingRule.expect("[INFO] Uploading file '/path' to transport request '002'.")
        loggingRule.expect("[INFO] File '/path' has been successfully uploaded to transport request '002'.")

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void uploadFileToTransportRequestCTS(
                                              String transportRequestId,
                                              String filePath,
                                              String endpoint,
                                              String credentialsId,
                                              String cmclientOpts) {

                cmUtilReceivedParams.transportRequestId = transportRequestId
                cmUtilReceivedParams.filePath = filePath
                cmUtilReceivedParams.endpoint = endpoint
                cmUtilReceivedParams.credentialsId = credentialsId
                cmUtilReceivedParams.cmclientOpts = cmclientOpts
            }
        }

        stepRule.step.transportRequestUploadFile(script: nullScript,
                      changeManagement: [type: 'CTS'],
                      transportRequestId: '002',
                      filePath: '/path',
                      cmUtils: cm)

        assert cmUtilReceivedParams ==
            [
                transportRequestId: '002',
                filePath: '/path',
                endpoint: 'https://example.org/cm',
                credentialsId: 'CM',
                cmclientOpts: ''
            ]
    }

    @Test
    public void uploadFileToTransportRequestRFCSanityChecksTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage(allOf(
            containsString('NO VALUE AVAILABLE FOR'),
            containsString('applicationUrl'),
            containsString('developmentInstance'),
            containsString('developmentClient'),
            containsString('applicationDescription'),
            containsString('abapPackage'),
            containsString('applicationName')))

        stepRule.step.transportRequestUploadFile(script: nullScript,
                 transportRequestId: '123456', //no sanity check, can be read from git history
                 changeManagement: [type: 'RFC'],
        )
    }

    @Test
    public void uploadFileToTransportRequestRFCSuccessTest() {

        def cmUtilsReceivedParams

        nullScript.commonPipelineEnvironment.configuration =
        [general:
            [changeManagement:
                [
                 endpoint: 'https://example.org/rfc'
                ]
            ]
        ]

        def cm = new ChangeManagement(nullScript) {

            void uploadFileToTransportRequestRFC(
                Map docker,
                String transportRequestId,
                String applicationId,
                String applicationURL,
                String endpoint,
                String credentialsId,
                String developmentInstance,
                String developmentClient,
                String applicationDescription,
                String abapPackage,
                String codePage,
                boolean acceptUnixStyleLineEndings,
                boolean failUploadOnWarning,
                boolean verbose) {

                cmUtilsReceivedParams = [
                    docker: docker,
                    transportRequestId: transportRequestId,
                    applicationName: applicationId,
                    applicationURL: applicationURL,
                    endpoint: endpoint,
                    credentialsId: credentialsId,
                    developmentInstance: developmentInstance,
                    developmentClient: developmentClient,
                    applicationDescription: applicationDescription,
                    abapPackage: abapPackage,
                    codePage: codePage,
                    acceptUnixStyleLineEndings: acceptUnixStyleLineEndings,
                    failUploadOnWarning: failUploadOnWarning,
                ]
            }
        }

        stepRule.step.transportRequestUploadFile(script: nullScript,
                 applicationUrl: 'http://example.org/blobstore/xyz.zip',
                 codePage: 'UTF-9',
                 acceptUnixStyleLineEndings: true,
                 transportRequestId: '123456',
                 changeManagement: [
                     type: 'RFC',
                     rfc: [
                         developmentClient: '002',
                         developmentInstance: '001'
                     ]
                 ],
                 applicationName: '42',
                 applicationDescription: 'Lorem ipsum',
                 abapPackage: 'XYZ',
                 cmUtils: cm,)

        assert cmUtilsReceivedParams ==
            [
                docker: [
                    image: 'rfc',
                    options: [],
                    envVars: [:],
                    imagePull: true
                ],
                transportRequestId: '123456',
                applicationName: '42',
                applicationURL: 'http://example.org/blobstore/xyz.zip',
                endpoint: 'https://example.org/rfc',
                credentialsId: 'CM',
                developmentInstance: '001',
                developmentClient: '002',
                applicationDescription: 'Lorem ipsum',
                abapPackage:'XYZ',
                codePage: 'UTF-9',
                acceptUnixStyleLineEndings: true,
                failUploadOnWarning: true,
            ]
    }

    @Test
    public void uploadFileToTransportRequestRFCUploadFailsTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('upload failed')

        def cm = new ChangeManagement(nullScript) {

            void uploadFileToTransportRequestRFC(
                Map docker,
                String transportRequestId,
                String applicationId,
                String applicationURL,
                String endpoint,
                String credentialsId,
                String developmentInstance,
                String developmentClient,
                String applicationDescription,
                String abapPackage,
                String codePage,
                boolean acceptUnixStyleLineEndings,
                boolean failOnUploadWarning,
                boolean verbose) {
                throw new ChangeManagementException('upload failed')
            }
        }

        stepRule.step.transportRequestUploadFile(script: nullScript,
                 applicationUrl: 'http://example.org/blobstore/xyz.zip',
                 codePage: 'UTF-9',
                 acceptUnixStyleLineEndings: true,
                 transportRequestId: '123456',
                 changeManagement: [
                     type: 'RFC',
                     rfc: [
                         docker: [
                             image: 'rfc',
                             options: [],
                             envVars: [:],
                             imagePull: false,
                         ],
                         developmentClient: '002',
                         developmentInstance: '001',
                         ]
                     ],
                 applicationName: '42',
                 applicationDescription: 'Lorem ipsum',
                 abapPackage: 'XYZ',
                 cmUtils: cm,)
    }

    @Test
    public void uploadFileToTransportRequestSOLMANSuccessTest() {

        // Here we test only the case where the transportRequestId is
        // provided via parameters. The other cases are tested by
        // corresponding tests for StepHelpers#getTransportRequestId(./.)

        loggingRule.expect("[INFO] Uploading file '/path' to transport request '002' of change document '001'.")
        loggingRule.expect("[INFO] File '/path' has been successfully uploaded to transport request '002' of change document '001'.")

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void uploadFileToTransportRequestSOLMAN(
                                              String changeId,
                                              String transportRequestId,
                                              String applicationId,
                                              String filePath,
                                              String endpoint,
                                              String credentialsId,
                                              String cmclientOpts) {

                cmUtilReceivedParams.changeId = changeId
                cmUtilReceivedParams.transportRequestId = transportRequestId
                cmUtilReceivedParams.applicationId = applicationId
                cmUtilReceivedParams.filePath = filePath
                cmUtilReceivedParams.endpoint = endpoint
                cmUtilReceivedParams.credentialsId = credentialsId
                cmUtilReceivedParams.cmclientOpts = cmclientOpts
            }
        }

        stepRule.step.transportRequestUploadFile(script: nullScript,
                      changeDocumentId: '001',
                      transportRequestId: '002',
                      applicationId: 'app',
                      filePath: '/path',
                      cmUtils: cm)

        assert cmUtilReceivedParams ==
            [
                changeId: '001',
                transportRequestId: '002',
                applicationId: 'app',
                filePath: '/path',
                endpoint: 'https://example.org/cm',
                credentialsId: 'CM',
                cmclientOpts: ''
            ]
    }

    @Test
    public void uploadFileToTransportRequestSOLMANSuccessApplicationIdFromConfigurationTest() {

        nullScript.commonPipelineEnvironment.configuration.put(['steps',
                                                                   [transportRequestUploadFile:
                                                                       [applicationId: 'AppIdfromConfig']]])

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void uploadFileToTransportRequestSOLMAN(
                                              String changeId,
                                              String transportRequestId,
                                              String applicationId,
                                              String filePath,
                                              String endpoint,
                                              String credentialsId,
                                              String cmclientOpts) {

                cmUtilReceivedParams.applicationId = applicationId
            }
        }

        stepRule.step.transportRequestUploadFile(
                      script: nullScript,
                      changeDocumentId: '001',
                      transportRequestId: '002',
                      filePath: '/path',
                      cmUtils: cm)

        assert cmUtilReceivedParams.applicationId == 'AppIdfromConfig'
    }

    @Test
    public void uploadFileToTransportRequestSOLMANFilePathFromParameters() {

        // this one is not used when file path is provided via signature
        nullScript.commonPipelineEnvironment.setMtarFilePath('/path2')

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void uploadFileToTransportRequestSOLMAN(
                                              String changeId,
                                              String transportRequestId,
                                              String applicationId,
                                              String filePath,
                                              String endpoint,
                                              String credentialsId,
                                              String cmclientOpts) {

                cmUtilReceivedParams.filePath = filePath
            }
        }

        stepRule.step.transportRequestUploadFile(script: nullScript,
                      changeDocumentId: '001',
                      transportRequestId: '002',
                      applicationId: 'app',
                      filePath: '/path',
                      cmUtils: cm)

        assert cmUtilReceivedParams.filePath == '/path'
    }

    @Test
    public void uploadFileToTransportRequestSOLMANFilePathFromCommonPipelineEnvironment() {

        // this one is used since there is nothing in the signature
        nullScript.commonPipelineEnvironment.setMtarFilePath('/path2')

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void uploadFileToTransportRequestSOLMAN(
                                              String changeId,
                                              String transportRequestId,
                                              String applicationId,
                                              String filePath,
                                              String endpoint,
                                              String credentialsId,
                                              String cmclientOpts) {

                cmUtilReceivedParams.filePath = filePath
            }
        }

        stepRule.step.transportRequestUploadFile(script: nullScript,
                      changeDocumentId: '001',
                      transportRequestId: '002',
                      applicationId: 'app',
                      cmUtils: cm)

        assert cmUtilReceivedParams.filePath == '/path2'
    }

    @Test
    public void uploadFileToTransportRequestSOLMANUploadFailureTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('Upload failure.')

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void uploadFileToTransportRequestSOLMAN(
                                              String changeId,
                                              String transportRequestId,
                                              String applicationId,
                                              String filePath,
                                              String endpoint,
                                              String credentialsId,
                                              String cmclientOpts) {
                throw new ChangeManagementException('Upload failure.')
            }
        }

        stepRule.step.transportRequestUploadFile(script: nullScript,
                      changeDocumentId: '001',
                      transportRequestId: '001',
                      applicationId: 'app',
                      filePath: '/path',
                      cmUtils: cm)
    }

    @Test
    public void invalidBackendTypeTest() {
        thrown.expect(AbortException)
        thrown.expectMessage('Invalid backend type: \'DUMMY\'. Valid values: [SOLMAN, CTS, RFC, NONE]. ' +
                             'Configuration: \'changeManagement/type\'.')

        stepRule.step.transportRequestUploadFile(script: nullScript,
                      applicationId: 'app',
                      filePath: '/path',
                      changeManagement: [type: 'DUMMY'])

    }

    @Test
    public void cmIntegrationSwichtedOffTest() {

        loggingRule.expect('[INFO] Change management integration intentionally switched off.')

        stepRule.step.transportRequestUploadFile(script: nullScript,
            changeManagement: [type: 'NONE'])
    }

}
