import java.util.Map

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

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
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(thrown)
        .around(new JenkinsReadYamlRule(this))
        .around(jsr)
        .around(jlr)
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
                                          endpoint: 'https://example.org/cm'
                                         ]
                                     ]
                                 ]
    }

    @Test
    public void changeDocumentIdNotProvidedTest() {

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

        jsr.step.call(script: nullScript, transportRequestId: '001', applicationId: 'app', filePath: '/path', cmUtils: cm)
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

        jsr.step.call(script: nullScript, changeDocumentId: '001', applicationId: 'app', filePath: '/path', cmUtils: cm)
    }

    @Test
    public void applicationIdNotProvidedTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("ERROR - NO VALUE AVAILABLE FOR applicationId")

        jsr.step.call(script: nullScript, changeDocumentId: '001', transportRequestId: '001', filePath: '/path')
    }

    @Test
    public void filePathNotProvidedTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("ERROR - NO VALUE AVAILABLE FOR filePath")

        jsr.step.call(script: nullScript, changeDocumentId: '001', transportRequestId: '001', applicationId: 'app')
    }

    @Test
    public void uploadFileToTransportRequestFailureTest() {

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void uploadFileToTransportRequest(String changeId,
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

        jsr.step.call(script: nullScript,
                      changeDocumentId: '001',
                      transportRequestId: '001',
                      applicationId: 'app',
                      filePath: '/path',
                      cmUtils: cm)
    }

    @Test
    public void uploadFileToTransportRequestSuccessTest() {

        jlr.expect("[INFO] Uploading file '/path' to transport request '002' of change document '001'.")
        jlr.expect("[INFO] File '/path' has been successfully uploaded to transport request '002' of change document '001'.")

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void uploadFileToTransportRequest(String changeId,
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

        jsr.step.call(script: nullScript,
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
    public void uploadFileToTransportRequestSuccessApplicationIdFromConfigurationTest() {

        nullScript.commonPipelineEnvironment.configuration.put(['steps',
                                                                   [transportRequestUploadFile:
                                                                       [applicationId: 'AppIdfromConfig']]])

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void uploadFileToTransportRequest(String changeId,
                                              String transportRequestId,
                                              String applicationId,
                                              String filePath,
                                              String endpoint,
                                              String credentialsId,
                                              String cmclientOpts) {

                cmUtilReceivedParams.applicationId = applicationId
            }
        }

        jsr.step.transportRequestUploadFile(
                      script: nullScript,
                      changeDocumentId: '001',
                      transportRequestId: '002',
                      filePath: '/path',
                      cmUtils: cm)

        assert cmUtilReceivedParams.applicationId == 'AppIdfromConfig'
    }

    @Test
    public void uploadFileToTransportRequestFilePathFromParameters() {

        // this one is not used when file path is provided via signature
        nullScript.commonPipelineEnvironment.setMtarFilePath('/path2')

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void uploadFileToTransportRequest(String changeId,
                                              String transportRequestId,
                                              String applicationId,
                                              String filePath,
                                              String endpoint,
                                              String credentialsId,
                                              String cmclientOpts) {

                cmUtilReceivedParams.filePath = filePath
            }
        }

        jsr.step.call(script: nullScript,
                      changeDocumentId: '001',
                      transportRequestId: '002',
                      applicationId: 'app',
                      filePath: '/path',
                      cmUtils: cm)

        assert cmUtilReceivedParams.filePath == '/path'
    }

    @Test
    public void uploadFileToTransportRequestFilePathFromCommonPipelineEnvironment() {

        // this one is used since there is nothing in the signature
        nullScript.commonPipelineEnvironment.setMtarFilePath('/path2')

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void uploadFileToTransportRequest(String changeId,
                                              String transportRequestId,
                                              String applicationId,
                                              String filePath,
                                              String endpoint,
                                              String credentialsId,
                                              String cmclientOpts) {

                cmUtilReceivedParams.filePath = filePath
            }
        }

        jsr.step.call(script: nullScript,
                      changeDocumentId: '001',
                      transportRequestId: '002',
                      applicationId: 'app',
                      cmUtils: cm)

        assert cmUtilReceivedParams.filePath == '/path2'
    }

    @Test
    public void uploadFileToTransportRequestUploadFailureTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('Upload failure.')

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void uploadFileToTransportRequest(String changeId,
                                              String transportRequestId,
                                              String applicationId,
                                              String filePath,
                                              String endpoint,
                                              String credentialsId,
                                              String cmclientOpts) {
                throw new ChangeManagementException('Upload failure.')
            }
        }

        jsr.step.call(script: nullScript,
                      changeDocumentId: '001',
                      transportRequestId: '001',
                      applicationId: 'app',
                      filePath: '/path',
                      cmUtils: cm)
    }

}
