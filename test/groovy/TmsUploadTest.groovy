import com.sap.piper.integration.TransportManagementService
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.junit.Assert.assertThat

public class TmsUploadTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsEnvironmentRule envRule = new JenkinsEnvironmentRule(this)

    def tmsStub
    def calledTmsMethodsWithArgs = []
    def uri = "https://dummy-url.com"
    def uaaUrl = "https://oauth.com"
    def oauthClientId = "myClientId"
    def oauthClientSecret = "myClientSecret"
    def serviceKeyContent = """{ 
                                "uri": "${uri}",
                                "uaa": {
                                    "clientid": "${oauthClientId}",
                                    "clientsecret": "${oauthClientSecret}",
                                    "url": "${uaaUrl}"
                                }
                               }
                             """


    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(thrown)
        .around(new JenkinsReadYamlRule(this))
        .around(stepRule)
        .around(loggingRule)
        .around(envRule)
        .around(new JenkinsCredentialsRule(this)
            .withCredentials('TMS_ServiceKey', serviceKeyContent))

    @Before
    public void setup() {
        tmsStub = mockTransportManagementService()
        helper.registerAllowedMethod("unstash", [String.class], { s -> return [s] })
    }

    @After
    void tearDown() {
        calledTmsMethodsWithArgs.clear()
    }

    @Test
    public void minimalConfig__isSuccessful() {
        def userIdCause = new hudson.model.Cause.UserIdCause()
        userIdCause.metaClass.getUserId =  {
            return "Test User"
        }
        binding.currentBuild.getRawBuild = {
            return [getCauses: {
                return [userIdCause]
            }]
        }
        binding.workspace = "."
        envRule.env.gitCommitId = "testCommitId"

        stepRule.step.tmsUpload(
            script: nullScript,
            juStabUtils: utils,
            transportManagementService: tmsStub,
            mtaPath: 'dummy.mtar',
            nodeName: 'myNode',
            credentialsId: 'TMS_ServiceKey'
        )

        assertThat(calledTmsMethodsWithArgs[0], is("authentication('${uaaUrl}', '${oauthClientId}', '${oauthClientSecret}')"))
        assertThat(calledTmsMethodsWithArgs[1], is("uploadFileToTMS('${uri}', 'myToken', './dummy.mtar', 'Test User')"))
        assertThat(calledTmsMethodsWithArgs[2], is("uploadFileToNode('${uri}', 'myToken', 'myNode', '1234', 'Git CommitId: testCommitId')"))
        assertThat(loggingRule.log, containsString("[TransportManagementService] File './dummy.mtar' successfully uploaded to Node 'myNode' (Id: '1000')."))
        assertThat(loggingRule.log, containsString("[TransportManagementService] Corresponding Transport Request: 'Git CommitId: testCommitId' (Id: '2000')"))
        assertThat(loggingRule.log, not(containsString("[TransportManagementService] CredentialsId: 'TMS_ServiceKey'")))

    }

    @Test
    public void verboseMode__yieldsMoreEchos() {
        def userIdCause = new hudson.model.Cause.UserIdCause()
        userIdCause.metaClass.getUserId =  {
            return "Test User"
        }
        binding.currentBuild.getRawBuild = {
            return [getCauses: {
                return [userIdCause]
            }]
        }
        binding.workspace = "."
        envRule.env.gitCommitId = "testCommitId"

        stepRule.step.tmsUpload(
            script: nullScript,
            juStabUtils: utils,
            transportManagementService: tmsStub,
            mtaPath: 'dummy.mtar',
            nodeName: 'myNode',
            credentialsId: 'TMS_ServiceKey',
            verbose: true
        )

        assertThat(loggingRule.log, containsString("[TransportManagementService] CredentialsId: 'TMS_ServiceKey'"))
        assertThat(loggingRule.log, containsString("[TransportManagementService] Node name: 'myNode'"))
        assertThat(loggingRule.log, containsString("[TransportManagementService] MTA path: 'dummy.mtar'"))
        assertThat(loggingRule.log, containsString("[TransportManagementService] Named user: 'Test User'"))
        assertThat(loggingRule.log, containsString("[TransportManagementService] UAA URL: '${uaaUrl}'"))
        assertThat(loggingRule.log, containsString("[TransportManagementService] TMS URL: '${uri}'"))
        assertThat(loggingRule.log, containsString("[TransportManagementService] ClientId: '${oauthClientId}'"))
    }

    @Test
    public void noUserAvailableInCurrentBuild__usesDefaultUser() {
        binding.currentBuild.getRawBuild = {
            return [getCauses: {
                return []
            }]
        }
        binding.workspace = "."
        envRule.env.gitCommitId = "testCommitId"

        stepRule.step.tmsUpload(
            script: nullScript,
            juStabUtils: utils,
            transportManagementService: tmsStub,
            mtaPath: 'dummy.mtar',
            nodeName: 'myNode',
            credentialsId: 'TMS_ServiceKey'
        )

        assertThat(calledTmsMethodsWithArgs[1], is("uploadFileToTMS('${uri}', 'myToken', './dummy.mtar', 'Piper-Pipeline')"))
    }

    @Test
    public void addCustomDescription__descriptionChanged() {
        def userIdCause = new hudson.model.Cause.UserIdCause()
        userIdCause.metaClass.getUserId =  {
            return "Test User"
        }
        binding.currentBuild.getRawBuild = {
            return [getCauses: {
                return [userIdCause]
            }]
        }
        binding.workspace = "."
        envRule.env.gitCommitId = "testCommitId"

        stepRule.step.tmsUpload(
            script: nullScript,
            juStabUtils: utils,
            transportManagementService: tmsStub,
            mtaPath: 'dummy.mtar',
            nodeName: 'myNode',
            credentialsId: 'TMS_ServiceKey',
            customDescription: 'My custom description for testing.'
        )

        assertThat(calledTmsMethodsWithArgs[2], is("uploadFileToNode('${uri}', 'myToken', 'myNode', '1234', 'My custom description for testing. Git CommitId: testCommitId')"))
        assertThat(loggingRule.log, containsString("[TransportManagementService] Corresponding Transport Request: 'My custom description for testing. Git CommitId: testCommitId' (Id: '2000')"))
    }

    def mockTransportManagementService() {
        return new TransportManagementService(nullScript, [:]) {
            def authentication(String uaaUrl, String oauthClientId, String oauthClientSecret) {
                calledTmsMethodsWithArgs << "authentication('${uaaUrl}', '${oauthClientId}', '${oauthClientSecret}')"
                return "myToken"
            }

            def uploadFileToTMS(String url, String token, String file, String namedUser) {
                calledTmsMethodsWithArgs << "uploadFileToTMS('${url}', '${token}', '${file}', '${namedUser}')"
                return [fileId: 1234, fileName: file]
            }

            def uploadFileToNode(String url, String token, String nodeName, int fileId, String description, String namedUser) {
                calledTmsMethodsWithArgs << "uploadFileToNode('${url}', '${token}', '${nodeName}', '${fileId}', '${description}')"
                return [transportRequestDescription: description, transportRequestId: 2000, queueEntries: [nodeName: 'myNode', nodeId: 1000]]
            }
        }
    }
}
