import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.Utils
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException
import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

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

        stepRule.step.transportRequestUploadFile(script: nullScript, transportRequestId: '001', applicationId: 'app', filePath: '/path', 
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
    public void transportRequestIdNotProvidedTest() {

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

        stepRule.step.transportRequestUploadFile(script: nullScript, changeDocumentId: '001', applicationId: 'app', filePath: '/path', 
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

        thrown.expect(AbortException)
        thrown.expectMessage("Exception message")

        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
                throw new AbortException('Exception message')
            }
        )

        stepRule.step.transportRequestUploadFile(script: nullScript,
                      changeDocumentId: '001',
                      transportRequestId: '001',
                      applicationId: 'app',
                      filePath: '/path',
                      changeManagement: [
                          type: 'SOLMAN',
                          endpoint: 'https://example.org/cm',
                          clientOpts: '--client opts'
                      ],
                      credentialsId: 'CM'
                  )
    }

    @Test
    public void uploadFileToTransportRequestCTSSuccessTest() {

        loggingRule.expect("[INFO] Uploading application 'myApp' to transport request '002'.")
        loggingRule.expect("[INFO] Application 'myApp' has been successfully uploaded to transport request '002'.")

        def calledWithParameters,
            calledWithStepName,
            calledWithMetadata,
            calledWithCredentials

        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
            params, stepName, metaData, creds ->
                calledWithParameters = params
                calledWithStepName = stepName
                calledWithMetadata = metaData
                calledWithCredentials = creds
            })

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
                      credentialsId: 'CM')

        assertThat(calledWithStepName, is('transportRequestUploadCTS'))
        assertThat(calledWithParameters.transportRequestId, is('002'))
        assertThat(calledWithParameters.endpoint, is('https://example.org/cm'))
        assertThat(calledWithParameters.client, is('001'))
        assertThat(calledWithParameters.applicationName, is('myApp'))
        assertThat(calledWithParameters.description, is('the description'))
        assertThat(calledWithParameters.abapPackage, is('myPackage'))
        assertThat(calledWithParameters.osDeployUser, is('node2'))
        assertThat(calledWithParameters.deployToolDependencies, is(['@ui5/cli', '@sap/ux-ui5-tooling', '@ui5/logger', '@ui5/fs', '@dummy/foo']))
        assertThat(calledWithParameters.npmInstallOpts, is(['--verbose']))
        assertThat(calledWithParameters.deployConfigFile, is('ui5-deploy.yaml'))
        assertThat(calledWithParameters.uploadCredentialsId, is('CM'))
    }

    @Test
    public void uploadFileToTransportRequestCTSDockerParams() {

        loggingRule.expect("[INFO] Uploading application 'myApp' to transport request '002'.")
        loggingRule.expect("[INFO] Application 'myApp' has been successfully uploaded to transport request '002'.")

        def calledWithParameters

        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
            params, stepName, metaData, creds ->
                calledWithParameters = params
            })

        stepRule.step.transportRequestUploadFile(script: nullScript,
                      changeManagement: [
                          type: 'CTS',
                          client: '001',
                          cts: [
                              osDeployUser: 'node2',
                              deployToolDependencies: ['@ui5/cli', '@sap/ux-ui5-tooling', '@ui5/logger', '@ui5/fs', '@dummy/foo'],
                              npmInstallOpts: ['--verbose'],
                              nodeDocker: [
                                   image: 'ctsImage',
                                   options: ['-o1', 'opt1', '-o2', 'opt2'],
                                   envVars: [env1: 'env1', env2: 'env2'],
                                   pullImage: false,
                              ],
                          ]
                      ],
                      applicationName: 'myApp',
                      applicationDescription: 'the description',
                      abapPackage: 'myPackage',
                      transportRequestId: '002',
                      credentialsId: 'CM')

        assertThat(calledWithParameters.dockerImage, is('ctsImage'))
        assertThat(calledWithParameters.dockerOptions, is(['-o1', 'opt1', '-o2', 'opt2']))
        assertThat(calledWithParameters.dockerEnvVars, is([env1: 'env1', env2: 'env2']))
        assertThat(calledWithParameters.dockerPullImage, is(false))
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

        def calledWithParameters,
            calledWithStepName,
            calledWithMetadata,
            calledWithCredentials

        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
            params, stepName, metaData, creds ->
                calledWithParameters = params
                calledWithStepName = stepName
                calledWithMetadata = metaData
                calledWithCredentials = creds
            }
        )

        nullScript.commonPipelineEnvironment.configuration =
        [general:
            [changeManagement:
                [
                 endpoint: 'https://example.org/rfc'
                ]
            ]
        ]

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
                 credentialsId: 'CM'
        )

        assertThat(calledWithStepName, is('transportRequestUploadRFC'))
        assertThat(calledWithParameters.applicationName, is('42'))
        assertThat(calledWithParameters.applicationUrl, is('http://example.org/blobstore/xyz.zip'))
        assertThat(calledWithParameters.endpoint, is('https://example.org/rfc'))
        assertThat(calledWithParameters.uploadCredentialsId, is('CM'))
        assertThat(calledWithParameters.instance, is('001'))
        assertThat(calledWithParameters.client, is('002'))
        assertThat(calledWithParameters.applicationDescription, is('Lorem ipsum'))
        assertThat(calledWithParameters.abapPackage, is('XYZ'))
        assertThat(calledWithParameters.codePage, is('UTF-9'))
        assertThat(calledWithParameters.acceptUnixStyleLineEndings, is(true))
        assertThat(calledWithParameters.failUploadOnWarning, is(true))
        assertThat(calledWithParameters.transportRequestId, is('123456'))
    }

    @Test
    public void uploadFileToTransportRequestRFCUploadFailsTest() {

        thrown.expect(AbortException)
        thrown.expectMessage('upload failed')

        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
                throw new AbortException('upload failed')
            }
        )

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
                 credentialsId: 'CM'
        )
    }

    @Test
    public void uploadFileToTransportRequestRFCDockerParams() {

        def calledWithParameters

        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
            params, stepName, metaData, creds ->
                calledWithParameters = params
            }
        )

        stepRule.step.transportRequestUploadFile(script: nullScript,
                 applicationUrl: 'http://example.org/blobstore/xyz.zip',
                 codePage: 'UTF-9',
                 acceptUnixStyleLineEndings: true,
                 transportRequestId: '123456',
                 changeManagement: [
                     type: 'RFC',
                     endpoint: 'https://example.org/cm',
                     rfc: [
                         developmentClient: '002',
                         developmentInstance: '001',
                         docker: [
                              image: 'rfcImage',
                              options: ['-o1', 'opt1', '-o2', 'opt2'],
                              envVars: [env1: 'env1', env2: 'env2'],
                              pullImage: false,
                         ],
                     ]
                 ],
                 applicationName: '42',
                 applicationDescription: 'Lorem ipsum',
                 abapPackage: 'XYZ',
                 credentialsId: 'CM'
        )

        assertThat(calledWithParameters.dockerImage, is('rfcImage'))
        assertThat(calledWithParameters.dockerOptions, is(['-o1', 'opt1', '-o2', 'opt2']))
        assertThat(calledWithParameters.dockerEnvVars, is([env1: 'env1', env2: 'env2']))
        assertThat(calledWithParameters.dockerPullImage, is(false))
    }

    @Test
    public void uploadFileToTransportRequestSOLMANSuccessTest() {

        loggingRule.expect("[INFO] Uploading file '/path' to transport request '002' of change document '001'.")
        loggingRule.expect("[INFO] File '/path' has been successfully uploaded to transport request '002' of change document '001'.")

        def calledWithParameters,
            calledWithStepName,
            calledWithMetadata,
            calledWithCredentials

        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
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
    public void uploadFileToTransportRequestSOLMANSuccessApplicationIdFromConfigurationTest() {

        def calledWithParameters,
            calledWithStepName,
            calledWithMetadata,
            calledWithCredentials

        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
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

        stepRule.step.transportRequestUploadFile(
                      script: nullScript,
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
    public void uploadFileToTransportRequestSOLMANFilePathFromParameters() {

        // this one is not used when file path is provided via signature
        nullScript.commonPipelineEnvironment.setMtarFilePath('/path2')

        def calledWithParameters,
            calledWithStepName,
            calledWithMetadata,
            calledWithCredentials

        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
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
    public void uploadFileToTransportRequestSOLMANFilePathFromCommonPipelineEnvironment() {

        // this one is used since there is nothing in the signature
        nullScript.commonPipelineEnvironment.setMtarFilePath('/path2')

        def calledWithParameters,
            calledWithStepName,
            calledWithMetadata,
            calledWithCredentials

        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
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

        assertThat(calledWithParameters.filePath, is('/path2'))
    }

    @Test
    public void uploadFileToTransportRequestSOLMANDockerParams() {

        // this one is used since there is nothing in the signature
        nullScript.commonPipelineEnvironment.setMtarFilePath('/path2')

        def calledWithParameters

        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
            params, stepName, metaData, creds ->
                calledWithParameters = params
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
                          clientOpts: '--client opts',
                          solman: [
                              docker: [
                                  image: 'solmanImage',
                                  options: ['-o1', 'opt1', '-o2', 'opt2'],
                                  envVars: [env1: 'env1', env2: 'env2'],
                                  pullImage: false,
                              ],
                          ],
                      ],
                      credentialsId: 'CM'
        )

        assertThat(calledWithParameters.dockerImage, is('solmanImage'))
        assertThat(calledWithParameters.dockerOptions, is(['-o1', 'opt1', '-o2', 'opt2']))
        assertThat(calledWithParameters.dockerEnvVars, is([env1: 'env1', env2: 'env2']))
        assertThat(calledWithParameters.dockerPullImage, is(false))
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
