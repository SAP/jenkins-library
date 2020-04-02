import com.sap.piper.JenkinsUtils
import com.sap.piper.integration.TransportManagementService

import hudson.AbortException

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
    private JenkinsFileExistsRule fileExistsRules = new JenkinsFileExistsRule(this, ['dummy.mtar', 'dummy.mtaext'])

    def tmsStub
    def jenkinsUtilsStub
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

    class JenkinsUtilsMock extends JenkinsUtils {
        def userId

        JenkinsUtilsMock(userId) {
            this.userId = userId
        }

        def getJobStartedByUserId(){
            return this.userId
        }
    }

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(thrown)
        .around(new JenkinsReadYamlRule(this))
        .around(stepRule)
        .around(loggingRule)
        .around(envRule)
        .around(fileExistsRules)
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
        jenkinsUtilsStub = new JenkinsUtilsMock("Test User")
        binding.workspace = "."
        envRule.env.gitCommitId = "testCommitId"

        stepRule.step.tmsUpload(
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtilsStub,
            transportManagementService: tmsStub,
            mtaPath: 'dummy.mtar',
            nodeName: 'myNode',
            credentialsId: 'TMS_ServiceKey'
        )

        assertThat(calledTmsMethodsWithArgs[0], is("authentication('${uaaUrl}', '${oauthClientId}', '${oauthClientSecret}')"))
        assertThat(calledTmsMethodsWithArgs[1], is("uploadFile('${uri}', 'myToken', './dummy.mtar', 'Test User')"))
        assertThat(calledTmsMethodsWithArgs[2], is("uploadFileToNode('${uri}', 'myToken', 'myNode', '1234', 'Git CommitId: testCommitId')"))
        assertThat(loggingRule.log, containsString("[TransportManagementService] File './dummy.mtar' successfully uploaded to Node 'myNode' (Id: '1000')."))
        assertThat(loggingRule.log, containsString("[TransportManagementService] Corresponding Transport Request: 'Git CommitId: testCommitId' (Id: '2000')"))
        assertThat(loggingRule.log, not(containsString("[TransportManagementService] CredentialsId: 'TMS_ServiceKey'")))

    }

    @Test
    public void verboseMode__yieldsMoreEchos() {
        jenkinsUtilsStub = new JenkinsUtilsMock("Test User")
        binding.workspace = "."
        envRule.env.gitCommitId = "testCommitId"

        stepRule.step.tmsUpload(
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtilsStub,
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
        jenkinsUtilsStub = new JenkinsUtilsMock(null)
        binding.workspace = "."
        envRule.env.gitCommitId = "testCommitId"

        stepRule.step.tmsUpload(
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtilsStub,
            transportManagementService: tmsStub,
            mtaPath: 'dummy.mtar',
            nodeName: 'myNode',
            credentialsId: 'TMS_ServiceKey'
        )

        assertThat(calledTmsMethodsWithArgs[1], is("uploadFile('${uri}', 'myToken', './dummy.mtar', 'Piper-Pipeline')"))
    }

    @Test
    public void addCustomDescription__descriptionChanged() {
        jenkinsUtilsStub = new JenkinsUtilsMock("Test User")
        binding.workspace = "."
        envRule.env.gitCommitId = "testCommitId"

        stepRule.step.tmsUpload(
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtilsStub,
            transportManagementService: tmsStub,
            mtaPath: 'dummy.mtar',
            nodeName: 'myNode',
            credentialsId: 'TMS_ServiceKey',
            customDescription: 'My custom description for testing.'
        )

        assertThat(calledTmsMethodsWithArgs[2], is("uploadFileToNode('${uri}', 'myToken', 'myNode', '1234', 'My custom description for testing.')"))
        assertThat(loggingRule.log, containsString("[TransportManagementService] Corresponding Transport Request: 'My custom description for testing.' (Id: '2000')"))
    }
	
	@Test
	public void addMtaExtensionDescriptor__isSuccessful() {
		List<Long> nodeIds = [ 1 ]
		
		jenkinsUtilsStub = new JenkinsUtilsMock("Test User")
		binding.workspace = "."
		envRule.env.gitCommitId = "testCommitId"

		stepRule.step.tmsUpload(
			script: nullScript,
			juStabUtils: utils,
			jenkinsUtilsStub: jenkinsUtilsStub,
			transportManagementService: tmsStub,
			mtaPath: 'dummy.mtar',
			nodeName: 'myNode',
			credentialsId: 'TMS_ServiceKey',
			mtaExtDescriptorPath: 'dummy.mtaext',
			mtaVersion: '0.0.1',
			nodeIds: nodeIds,
		)

		assertThat(loggingRule.log, containsString("[TransportManagementService] MTA Extention Descriptor './dummy.mtaext' (Id: '1') successfully uploaded to Node with id '1'."))
		assertThat(calledTmsMethodsWithArgs[3], is("uploadMtaExtDescriptorToNode('${uri}', 'myToken', '1', './dummy.mtaext', '0.0.1', 'Git CommitId: testCommitId', 'Test User')"))
	}

    @Test
    public void failOnMissingMtaFile() {

        thrown.expect(AbortException)
        thrown.expectMessage('Mta file \'dummy.mtar\' does not exist.')

        fileExistsRules.existingFiles.remove('dummy.mtar')
        jenkinsUtilsStub = new JenkinsUtilsMock("Test User")

        stepRule.step.tmsUpload(
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtilsStub,
            transportManagementService: tmsStub,
            mtaPath: 'dummy.mtar',
            nodeName: 'myNode',
            credentialsId: 'TMS_ServiceKey',
            customDescription: 'My custom description for testing.'
        )
    }
	
	@Test
	public void failOnMissingMtaExtDescriptorFile() {

		thrown.expect(AbortException)
		thrown.expectMessage('Mta extension descriptor file \'dummy.mtaext\' does not exist.')

		fileExistsRules.existingFiles.remove('dummy.mtaext')
		jenkinsUtilsStub = new JenkinsUtilsMock("Test User")

		stepRule.step.tmsUpload(
			script: nullScript,
			juStabUtils: utils,
			jenkinsUtilsStub: jenkinsUtilsStub,
			transportManagementService: tmsStub,
			mtaPath: 'dummy.mtar',
			nodeName: 1,
			credentialsId: 'TMS_ServiceKey',
			mtaExtDescriptorPath: 'dummy.mtaext',
			mtaVersion: '0.0.1',
			nodeIds: [1],
		)
	}
	
	@Test
	public void failOnEmptyNodeIdList() {

		thrown.expect(AbortException)
		thrown.expectMessage('List of Node id should not be empty.')

        jenkinsUtilsStub = new JenkinsUtilsMock("Test User")
        binding.workspace = "."
        envRule.env.gitCommitId = "testCommitId"

        stepRule.step.tmsUpload(
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtilsStub,
            transportManagementService: tmsStub,
            mtaPath: 'dummy.mtar',
            nodeName: 'myNode',
            credentialsId: 'TMS_ServiceKey',
			mtaExtDescriptorPath: 'dummy.mtaext',
			mtaVersion: '0.0.1'
        )
	}

    def mockTransportManagementService() {
        return new TransportManagementService(nullScript, [:]) {
            def authentication(String uaaUrl, String oauthClientId, String oauthClientSecret) {
                calledTmsMethodsWithArgs << "authentication('${uaaUrl}', '${oauthClientId}', '${oauthClientSecret}')"
                return "myToken"
            }

            def uploadFile(String url, String token, String file, String namedUser) {
                calledTmsMethodsWithArgs << "uploadFile('${url}', '${token}', '${file}', '${namedUser}')"
                return [fileId: 1234, fileName: file]
            }

            def uploadFileToNode(String url, String token, String nodeName, int fileId, String description, String namedUser) {
                calledTmsMethodsWithArgs << "uploadFileToNode('${url}', '${token}', '${nodeName}', '${fileId}', '${description}')"
                return [transportRequestDescription: description, transportRequestId: 2000, queueEntries: [nodeName: 'myNode', nodeId: 1000]]
            }

			def uploadMtaExtDescriptorToNode(String url, String token, Long nodeId, String file, String mtaVersion, String description, String namedUser) {
				calledTmsMethodsWithArgs << "uploadMtaExtDescriptorToNode('${url}', '${token}', '${nodeId}', '${file}', '${mtaVersion}', '${description}', '${namedUser}')"
				return [fileId: 1, fileName: file]
			}
        }
    }
}
