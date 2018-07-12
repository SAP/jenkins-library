import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import util.BasePiperTest
import util.JenkinsStepRule
import util.JenkinsLoggingRule
import util.Rules

import hudson.AbortException


public class TransportRequestUploadFileTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(thrown)
        .around(jsr)
        .around(jlr)

    @Before
    public void setup() {

        helper.registerAllowedMethod('usernamePassword', [Map.class], {m -> return m})

        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->

            credentialsId = l[0].credentialsId
            binding.setProperty('username', 'anonymous')
            binding.setProperty('password', '********')
            try {
                c()
            } finally {
                binding.setProperty('username', null)
                binding.setProperty('password', null)
            }
         })

        helper.registerAllowedMethod('sh', [Map], { Map m -> return 0 })

        nullScript.commonPipelineEnvironment.configuration = [steps:
                                     [transportRequestUploadFile:
                                         [
                                          credentialsId: 'CM',
                                          endpoint: 'https://example.org/cm'
                                         ]
                                     ]
                                 ]
    }

    @Test
    public void changeDocumentIdNotProvidedTest() {

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

        thrown.expect(AbortException)
        thrown.expectMessage("Change document id not provided (parameter: 'changeDocumentId' or via commit history).")

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

        thrown.expect(AbortException)
        thrown.expectMessage("Transport Request id not provided (parameter: 'transportRequestId' or via commit history).")

        jsr.step.call(script: nullScript, changeDocumentId: '001', applicationId: 'app', filePath: '/path', cmUtils: cm)
    }

    @Test
    public void applicationIdNotProvidedTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("Application id not provided (parameter: 'applicationId').")

        jsr.step.call(script: nullScript, changeDocumentId: '001', transportRequestId: '001', filePath: '/path')
    }

    @Test
    public void filePathNotProvidedTest() {

        thrown.expect(AbortException)
        thrown.expectMessage("File path not provided (parameter: 'filePath').")

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
                                              String username,
                                              String password,
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

        helper.registerAllowedMethod('sh', [Map], { Map m -> return 0 })

        jsr.step.call(script: nullScript, changeDocumentId: '001', transportRequestId: '001', applicationId: 'app', filePath: '/path')

        assert jlr.log.contains("[INFO] Uploading file '/path' to transport request '001' of change document '001'.")
        assert jlr.log.contains("[INFO] File '/path' has been successfully uploaded to transport request '001' of change document '001'.")
    }
}
