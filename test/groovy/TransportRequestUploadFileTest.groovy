
import java.util.List
import java.util.Map

import static org.hamcrest.Matchers.*
import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.hasEntry
import static org.junit.Assert.assertThat

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


    @Test
    public void trUploadFile_SOLMAN_uploadSucceeds_Test() {

        def calledWithParameters,
            calledWithStepName,
            calledWithMetadata,
            calledWithCredentials

        helper.registerAllowedMethod( 'piperExecuteBin', [Map, String, String, List], {
            params, stepName, metaData, creds ->
                calledWithParameters = params
                calledWithStepName = stepName
                calledWithMetadata = metaData
                calledWithCredentials = creds
            }
        )

        stepRule.step.transportRequestUploadFile(script: nullScript,
            changeDocumentId: '001',
            transportRequestId: '002',
            applicationId: 'app',
            filePath: '/path',
            changeManagement: [
                type: 'SOLMAN',
                endpoint: 'https://example.org/cm',
                clientOpts: '--client opts'
            ],
            credentialsId: 'CM'
        )

        assertThat(calledWithStepName, is('transportRequestUploadSOLMAN'))
        assertThat(calledWithParameters.changeDocumentId, is('001'))
        assertThat(calledWithParameters.transportRequestId, is('002'))
        assertThat(calledWithParameters.applicationId, is('app'))
        assertThat(calledWithParameters.filePath, is('/path'))
        assertThat(calledWithParameters.endpoint, is('https://example.org/cm'))
        assertThat(calledWithParameters.cmClientOpts, is('--client opts'))
        assertThat(calledWithParameters.uploadCredentialsId, is('CM'))
    }

    @Test
    public void trUploadFile_SOLMAN_paramFromStep_Test() {

        def calledWithParameters,
            calledWithStepName,
            calledWithMetadata,
            calledWithCredentials

        helper.registerAllowedMethod( 'piperExecuteBin', [Map, String, String, List], {
            params, stepName, metaData, creds ->
                calledWithParameters = params
                calledWithStepName = stepName
                calledWithMetadata = metaData
                calledWithCredentials = creds
            }
        )
        
        nullScript.commonPipelineEnvironment.configuration.put(['steps',
                                                                   [transportRequestUploadFile:
                                                                       [applicationId: 'AppIdfromConfig']]])

        stepRule.step.transportRequestUploadFile(script: nullScript,
            changeDocumentId: '001',
            transportRequestId: '002',
            filePath: '/path',
            changeManagement: [
                type: 'SOLMAN',
                endpoint: 'https://example.org/cm',
                clientOpts: '--client opts'
            ],
            credentialsId: 'CM'
        )

        assertThat(calledWithParameters.applicationId, is('AppIdfromConfig'))
    }
    
    @Test
    public void trUploadFile_SOLMAN_FilePathFromParameters_Test() {

        // this one is not used when file path is provided via signature
        nullScript.commonPipelineEnvironment.setMtarFilePath('/pathByCPE')

        def calledWithParameters,
            calledWithStepName,
            calledWithMetadata,
            calledWithCredentials

        helper.registerAllowedMethod( 'piperExecuteBin', [Map, String, String, List], {
            params, stepName, metaData, creds ->
                calledWithParameters = params
                calledWithStepName = stepName
                calledWithMetadata = metaData
                calledWithCredentials = creds
            }
        )
        
        stepRule.step.transportRequestUploadFile(script: nullScript,
            changeDocumentId: '001',
            transportRequestId: '002',
            applicationId: 'app',
            filePath: '/pathByParam',
            changeManagement: [
                type: 'SOLMAN',
                endpoint: 'https://example.org/cm',
                clientOpts: '--client opts'
            ],
            credentialsId: 'CM'
        )

        assertThat(calledWithParameters.filePath, is('/pathByParam'))
    }
    
    @Test
    public void trUploadFile_SOLMAN_FilePathFromCPE_Test() {

        // this one is not used when file path is provided via signature
        nullScript.commonPipelineEnvironment.setMtarFilePath('/pathByCPE')

        def calledWithParameters,
            calledWithStepName,
            calledWithMetadata,
            calledWithCredentials

        helper.registerAllowedMethod( 'piperExecuteBin', [Map, String, String, List], {
            params, stepName, metaData, creds ->
                calledWithParameters = params
                calledWithStepName = stepName
                calledWithMetadata = metaData
                calledWithCredentials = creds
            }
        )
        
        stepRule.step.transportRequestUploadFile(script: nullScript,
            changeDocumentId: '001',
            transportRequestId: '002',
            applicationId: 'app',
            changeManagement: [
                type: 'SOLMAN',
                endpoint: 'https://example.org/cm',
                clientOpts: '--client opts'
            ],
            credentialsId: 'CM'
        )

        assertThat(calledWithParameters.filePath, is('/pathByCPE'))
    }
    
    @Test
    public void trUploadFile_SOLMAN_failsIfAppidIsMissing_Test() {
        def calledWithParameters,
            calledWithStepName,
            calledWithMetadata,
            calledWithCredentials
    
        helper.registerAllowedMethod( 'piperExecuteBin', [Map, String, String, List], { 
            params, stepName, metaData, creds ->
                calledWithParameters = params
                calledWithStepName = stepName
                calledWithMetadata = metaData
                calledWithCredentials = creds
            }
        )
    
        // we expect the failure only for SOLMAN (which is the default).
        // Use case for CTS without applicationId is checked by the
        // straight forward test case for CTS
        
        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("ERROR - NO VALUE AVAILABLE FOR applicationId")

        stepRule.step.transportRequestUploadFile(script: nullScript,
            changeDocumentId: '001',
            transportRequestId: '002',
            filePath: '/pathByParam',
            changeManagement: [
                type: 'SOLMAN',
                endpoint: 'https://example.org/cm',
                clientOpts: '--client opts'
            ],
            credentialsId: 'CM'
        )
    }
    
    @Test
    public void trUploadFile_SOLMAN_failsIfDocidIsMissing_Test() {
      
        def calledWithParameters = null
    
        helper.registerAllowedMethod( 'piperExecuteBin', [Map, String, String, List],
            {
                params, stepName, metaData, creds -> 
                    if(stepName.equals("transportRequestDocIDFromGit")) {
                        calledWithParameters = params
                    }
            }
        )
        
        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("Change document id not provided (parameter: 'changeDocumentId' provided to the step call or via commit history).")

        stepRule.step.transportRequestUploadFile(script: nullScript,
            transportRequestId: '002',
            applicationId: 'app',
            filePath: '/pathByParam',
            changeManagement: [
                type: 'SOLMAN',
                endpoint: 'https://example.org/cm',
                clientOpts: '--client opts'
            ],
            credentialsId: 'CM'
        )
        
        assert calledWithParameters != null
    }
    
    @Test
    public void trUploadFile_SOLMAN_failsIfTridIsMissing_Test() {

        def calledWithParameters = null
    
        helper.registerAllowedMethod( 'piperExecuteBin', [Map, String, String, List],
            {
                params, stepName, metaData, creds -> 
                    if(stepName.equals("transportRequestReqIDFromGit")) {
                        calledWithParameters = params
                    }
            }
        )
        
        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("Transport request id not provided (parameter: 'transportRequestId' provided to the step call or via commit history).")

        stepRule.step.transportRequestUploadFile(script: nullScript,
            changeDocumentId: '001',
            applicationId: 'app',
            filePath: '/pathByParam',
            changeManagement: [
                type: 'SOLMAN',
                endpoint: 'https://example.org/cm',
                clientOpts: '--client opts'
            ],
            credentialsId: 'CM'
        )

        assert calledWithParameters != null
    }
    
    @Test
    public void trUploadFile_SOLMAN_failsIfFilePathIsMissingTest() {

        thrown.expect(IllegalArgumentException)
        thrown.expectMessage("ERROR - NO VALUE AVAILABLE FOR filePath")

        stepRule.step.transportRequestUploadFile(script: nullScript,
            changeDocumentId: '001',
            transportRequestId: '001',
            applicationId: 'app',
            changeManagement: [
                type: 'SOLMAN',
                endpoint: 'https://example.org/cm',
                clientOpts: '--client opts'
            ],
            credentialsId: 'CM'
            )
    }

    @Test
    public void trUploadFile_SOLMAN_failsIfStepThrowsException_Test() {
    
        helper.registerAllowedMethod( 'piperExecuteBin', [Map, String, String, List], { 
                throw new AbortException('piperExecuteBin throws exit code 1')
            }
        )
    
        thrown.expect(AbortException)
        thrown.expectMessage("piperExecuteBin throws exit code 1")

        stepRule.step.transportRequestUploadFile(script: nullScript,
            changeDocumentId: '001',
            transportRequestId: '002',
            applicationId: 'app',
            filePath: '/pathByParam',
            changeManagement: [
                type: 'SOLMAN',
                endpoint: 'https://example.org/cm',
                clientOpts: '--client opts'
            ],
            credentialsId: 'CM'
        )
    }


}
