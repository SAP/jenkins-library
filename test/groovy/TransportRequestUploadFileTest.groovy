import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import hudson.AbortException
import util.BasePiperTest
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


     static def expectedCredentialsConfigList = [[
         type: 'usernamePassword',
         id: 'CM',
         env: ['PIPER_username', 'PIPER_password'],
         resolveCredentialsId: false
     ]]

     def received = null

    @Before
    public void setup() {

        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], { m, stepName, metadataFilePath, credentials ->

            received = [
                stepName: stepName,
                metadataFilePath: metadataFilePath,
                credentials: credentials
            ]
        })

        nullScript.commonPipelineEnvironment.configuration = [general:
                                     [changeManagement:
                                         [
                                          credentialsId: 'CM'
                                         ]
                                     ]
                                 ]
    }

    @Test
    public void goCallTest() {

        def types = [
            'SOLMAN': [
                stepName:"transportRequestUploadSOLMAN",
                metadataFilePath: 'metadata/transportRequestUploadSOLMAN.yaml',
                credentials: expectedCredentialsConfigList,
                ],
            'CTS': [
                stepName:"transportRequestUploadCTS",
                metadataFilePath: 'metadata/transportRequestUploadCTS.yaml',
                credentials: expectedCredentialsConfigList,
                ],
            'RFC': [
                stepName:"transportRequestUploadRFC",
                metadataFilePath: 'metadata/transportRequestUploadRFC.yaml',
                credentials: expectedCredentialsConfigList,
                ],
        ]


        for (type in types) {
            def uploadType = type.key
            def expected = type.value
            stepRule.step.transportRequestUploadFile(script: nullScript,
                changeManagement: [type: uploadType])

            assert received == expected
        }
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
