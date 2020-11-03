import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString

import java.util.List
import java.util.Map

import org.hamcrest.Matchers
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import com.sap.piper.cm.BackendType
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsStepRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsDockerExecuteRule
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
        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
    }

    @Test
    public void changeDocumentIdNotProvidedSOLMANTest() {

        // we expect the failure only for SOLMAN (which is the default).
        // Use case for CTS without change document id is checked by the
        // straight forward test case for CTS

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("Change document id not provided (parameter: 'changeDocumentId' provided to the step call or via commit history).")

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
        thrown.expectMessage("Transport request id not provided (parameter: 'transportRequestId' provided to the step call or via commit history).")

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
                                              Map docker,
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

        loggingRule.expect("[INFO] Uploading application 'myApp' to transport request '002'.")
        loggingRule.expect("[INFO] Application 'myApp' has been successfully uploaded to transport request '002'.")

        ChangeManagement cm = new ChangeManagement(nullScript) {
            void uploadFileToTransportRequestCTS(
                                              Map docker,
                                              String transportRequestId,
                                              String endpoint,
                                              String client,
                                              String appName,
                                              String appDescription,
                                              String abapPackage,
                                              String osDeployUser,
                                              def deployToolsDependencies,
                                              def npmInstallArgs,
                                              String deployConfigFile,
                                              String credentialsId) {

                cmUtilReceivedParams.docker = docker
                cmUtilReceivedParams.transportRequestId = transportRequestId
                cmUtilReceivedParams.endpoint = endpoint
                cmUtilReceivedParams.client = client
                cmUtilReceivedParams.appName = appName
                cmUtilReceivedParams.appDescription = appDescription
                cmUtilReceivedParams.abapPackage = abapPackage
                cmUtilReceivedParams.osDeployUser = osDeployUser
                cmUtilReceivedParams.deployToolDependencies = deployToolsDependencies
                cmUtilReceivedParams.npmInstallOpts = npmInstallArgs
                cmUtilReceivedParams.deployConfigFile = deployConfigFile
                cmUtilReceivedParams.credentialsId = credentialsId
            }
        }

        stepRule.step.transportRequestUploadFile(script: nullScript,
                      changeManagement: [
                          type: 'CTS',
                          client: '001',
                          cts: [
                              osDeployUser: 'node2',
                              deployToolDependencies: ['@ui5/cli', '@sap/ux-ui5-tooling', '@ui5/logger', '@ui5/fs', '@dummy/foo'],
                              npmInstallOpts: ['--verbose'],
                          ]
                      ],
                      applicationName: 'myApp',
                      applicationDescription: 'the description',
                      abapPackage: 'myPackage',
                      transportRequestId: '002',
                      cmUtils: cm)

        assert cmUtilReceivedParams ==
            [
                docker: [
                    image: 'node',
                    options:[],
                    envVars:[:],
                    pullImage:true
                ],
                transportRequestId: '002',
                endpoint: 'https://example.org/cm',
                client: '001',
                appName: 'myApp',
                appDescription: 'the description',
                abapPackage: 'myPackage',
                osDeployUser: 'node2',
                deployToolDependencies: ['@ui5/cli', '@sap/ux-ui5-tooling', '@ui5/logger', '@ui5/fs', '@dummy/foo'],
                npmInstallOpts: ['--verbose'],
                deployConfigFile: 'ui5-deploy.yaml',
                credentialsId: 'CM',
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
                    pullImage: true
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
                             pullImage: false,
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
                                              Map docker,
                                              String changeId,
                                              String transportRequestId,
                                              String applicationId,
                                              String filePath,
                                              String endpoint,
                                              String credentialsId,
                                              String cmclientOpts) {

                cmUtilReceivedParams.docker = docker
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
                docker: [
                    image: 'ppiper/cm-client',
                    pullImage: true,
                    envVars: [:],
                    options: [],
                ],
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
                                              Map docker,
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
                                              Map docker,
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
                                              Map docker,
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
                                              Map docker,
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
