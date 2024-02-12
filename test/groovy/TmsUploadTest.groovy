import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import com.sap.piper.integration.TransportManagementService

import hudson.AbortException

import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*
import util.JenkinsReadYamlRule

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue

public class TmsUploadTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsEnvironmentRule envRule = new JenkinsEnvironmentRule(this)
    private JenkinsFileExistsRule fileExistsRules = new JenkinsFileExistsRule(this, ['dummy.mtar', 'mta.yaml', 'dummy.mtaext', 'dummy2.mtaext', 'invalidDummy.mtaext'])
    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this)

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
        .around(readYamlRule)
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
        readYamlRule.registerYaml("mta.yaml", new FileInputStream(new File("test/resources/TransportManagementService/mta.yaml")))
                    .registerYaml("dummy.mtaext", new FileInputStream(new File("test/resources/TransportManagementService/dummy.mtaext")))
                    .registerYaml("dummy2.mtaext", new FileInputStream(new File("test/resources/TransportManagementService/dummy2.mtaext")))
                    .registerYaml("invalidDummy.mtaext", new FileInputStream(new File("test/resources/TransportManagementService/invalidDummy.mtaext")))
        Utils.metaClass.echo = { def m -> }
    }

    @After
    void tearDown() {
        Utils.metaClass = null
        calledTmsMethodsWithArgs.clear()
    }

    @Test
    public void defaultUseGoStep__callsPiperExecuteBin() {
        String calledStep = ''
        String usedMetadataFile = ''
        List credInfo = []
        helper.registerAllowedMethod('piperExecuteBin', [Map, String, String, List], {
            Map parameters, String stepName,
            String metadataFile, List credentialInfo ->
                calledStep = stepName
                usedMetadataFile = metadataFile
                credInfo = credentialInfo
        })

        jenkinsUtilsStub = new JenkinsUtilsMock("Test User")

        stepRule.step.tmsUpload(
            script: nullScript,
            jenkinsUtilsStub: jenkinsUtilsStub,
            mtaPath: 'dummy.mtar',
            nodeName: 'myNode',
            credentialsId: 'TMS_ServiceKey'
        )

        assertEquals('tmsUpload', calledStep)
        assertEquals('metadata/tmsUpload.yaml', usedMetadataFile)

        // contains assertion does not work apparently when comparing a list of lists against an expected list
        boolean found = false
        credInfo.each { entry ->
            if (entry == [type: 'token', id: 'credentialsId', env: ['PIPER_serviceKey']]) {
                found = true
            }
        }
        assertTrue(found)
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
            credentialsId: 'TMS_ServiceKey',
            useGoStep: false
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
            verbose: true,
            useGoStep: false
        )

        assertThat(loggingRule.log, containsString("[TransportManagementService] Using deprecated Groovy implementation of 'tmsUpload' step instead of the default Golang one, since 'useGoStep' toggle parameter is explicitly set to 'false'."))
        assertThat(loggingRule.log, containsString("[TransportManagementService] WARNING: Note that the deprecated Groovy implementation will be completely removed after February 29th, 2024. Consider using the Golang implementation by not setting the 'useGoStep' toggle parameter to 'false'."))
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
            credentialsId: 'TMS_ServiceKey',
            useGoStep: false
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
            customDescription: 'My custom description for testing.',
            useGoStep: false
        )

        assertThat(calledTmsMethodsWithArgs[2], is("uploadFileToNode('${uri}', 'myToken', 'myNode', '1234', 'My custom description for testing.')"))
        assertThat(loggingRule.log, containsString("[TransportManagementService] Corresponding Transport Request: 'My custom description for testing.' (Id: '2000')"))
    }

    @Test
    public void uploadMtaExtensionDescriptor__isSuccessful() {
        Map nodeExtDescriptorMap = ["testNode1": "dummy.mtaext", "testNode2": "dummy2.mtaext"]

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
            nodeExtDescriptorMapping: nodeExtDescriptorMap,
            mtaVersion: '0.0.1',
            useGoStep: false
        )

        assertThat(loggingRule.log, containsString("[TransportManagementService] MTA Extension Descriptor with ID 'com.sap.piper.tms.test.extension' successfully uploaded to Node 'testNode1'."))
        assertThat(calledTmsMethodsWithArgs[3], is("uploadMtaExtDescriptorToNode('${uri}', 'myToken', 1, './dummy.mtaext', '0.0.1', 'Git CommitId: testCommitId', 'Test User')"))
        assertThat(loggingRule.log, containsString("[TransportManagementService] MTA Extension Descriptor with ID 'com.sap.piper.tms.test.another.extension' successfully uploaded to Node 'testNode2'."))
        assertThat(calledTmsMethodsWithArgs[5], is("uploadMtaExtDescriptorToNode('${uri}', 'myToken', 2, './dummy2.mtaext', '0.0.1', 'Git CommitId: testCommitId', 'Test User')"))
    }

    @Test
    public void updateMtaExtensionDescriptor__isSuccessful() {
        Map nodeExtDescriptorMap = ["testNode1": "dummy2.mtaext"]

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
            nodeExtDescriptorMapping: nodeExtDescriptorMap,
            mtaVersion: '1.2.2',
            useGoStep: false
        )

        assertThat(loggingRule.log, containsString("[TransportManagementService] MTA Extension Descriptor with ID 'com.sap.piper.tms.test.another.extension' successfully updated for Node 'testNode1'."))
        assertThat(calledTmsMethodsWithArgs[2], is("getMtaExtDescriptor('${uri}', 'myToken', 1, 'com.sap.piper.tms.test', '1.2.2')"))
        assertThat(calledTmsMethodsWithArgs[3], is("updateMtaExtDescriptor('${uri}', 'myToken', 1, 2, './dummy2.mtaext', '1.2.2', 'Git CommitId: testCommitId', 'Test User')"))
    }

    @Test
    public void testMtaBuildDescriptorFromCPE() {
        Map nodeExtDescriptorMap = ["testNode1": "dummy.mtaext"]

        jenkinsUtilsStub = new JenkinsUtilsMock("Test User")
        binding.workspace = "."
        envRule.env.gitCommitId = "testCommitId"

        nullScript.commonPipelineEnvironment.setValue("mtaBuildToolDesc","path/mta.yaml")
        readYamlRule.registerYaml('path/mta.yaml','ID: "com.sap.piper.tms.test"' + "\n" + 'version: "9.9.9"')
        fileExistsRules.existingFiles.add('path/mta.yaml')


        stepRule.step.tmsUpload(
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtilsStub,
            transportManagementService: tmsStub,
            mtaPath: 'dummy.mtar',
            nodeName: 'myNode',
            credentialsId: 'TMS_ServiceKey',
            nodeExtDescriptorMapping: nodeExtDescriptorMap,
            mtaVersion: '9.9.9',
            useGoStep: false
        )

        assertThat(calledTmsMethodsWithArgs[2], is("getMtaExtDescriptor('${uri}', 'myToken', 1, 'com.sap.piper.tms.test', '9.9.9')"))
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
            customDescription: 'My custom description for testing.',
            useGoStep: false
        )
    }

    @Test
    public void useMtaFilePathFromPipelineEnvironment() {
        jenkinsUtilsStub = new JenkinsUtilsMock("Test User")
        binding.workspace = "."
        envRule.env.gitCommitId = "testCommitId"
        envRule.env.mtarFilePath = 'dummy.mtar'

        stepRule.step.tmsUpload(
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtilsStub,
            transportManagementService: tmsStub,
            nodeName: 'myNode',
            credentialsId: 'TMS_ServiceKey',
            useGoStep: false
        )

        assertThat(calledTmsMethodsWithArgs[1], is("uploadFile('${uri}', 'myToken', './dummy.mtar', 'Test User')"))
        assertThat(loggingRule.log, containsString("[TransportManagementService] File './dummy.mtar' successfully uploaded to Node 'myNode' (Id: '1000')."))

    }

    @Test
    public void failOnMissingMtaYaml() {
        thrown.expect(AbortException)
        thrown.expectMessage("mta.yaml is not found in the root folder of the project.")

        Map nodeExtDescriptorMap = ["testNode1": "dummy.mtaext"]

        fileExistsRules.existingFiles.remove('mta.yaml')
        jenkinsUtilsStub = new JenkinsUtilsMock("Test User")

        stepRule.step.tmsUpload(
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtilsStub,
            transportManagementService: tmsStub,
            mtaPath: 'dummy.mtar',
            nodeName: 'myNode',
            credentialsId: 'TMS_ServiceKey',
            nodeExtDescriptorMapping: nodeExtDescriptorMap,
            mtaVersion: '0.0.1',
            useGoStep: false
        )
    }

    @Test
    public void failOnMissingIdAndVersionInMtaYaml() {
        thrown.expect(AbortException)
        thrown.expectMessage("Property 'ID' is not found in mta.yaml.")
        thrown.expectMessage("Property 'version' is not found in mta.yaml.")

        Map nodeExtDescriptorMap = ["testNode1": "dummy.mtaext"]

        readYamlRule.registerYaml("mta.yaml", "_schema-version: '3.1'")
        jenkinsUtilsStub = new JenkinsUtilsMock("Test User")

        stepRule.step.tmsUpload(
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtilsStub,
            transportManagementService: tmsStub,
            mtaPath: 'dummy.mtar',
            nodeName: 'myNode',
            credentialsId: 'TMS_ServiceKey',
            nodeExtDescriptorMapping: nodeExtDescriptorMap,
            mtaVersion: '0.0.1',
            useGoStep: false
        )
    }

    @Test
    public void failOnInvalidNodeExtDescriptorMapping() {
        thrown.expect(AbortException)
        thrown.expectMessage("MTA extension descriptor files [notexisted.mtaext, notexisted2.mtaext] don't exist.")
        thrown.expectMessage("Nodes [testNode3, testNode4] don't exist. Please check the node name or create these nodes.")
        thrown.expectMessage("Parameter [extends] in MTA extension descriptor files [invalidDummy.mtaext] is not the same as MTA ID.")

        // test on all kinds of errors: node doesn't exist, MTA ID in .mtaext is incorrect, and .mtaext file doesn't exist
        Map nodeExtDescriptorMap = ["testNode1": "invalidDummy.mtaext", "testNode3": "notexisted.mtaext", "testNode4": "notexisted2.mtaext"]

        jenkinsUtilsStub = new JenkinsUtilsMock("Test User")

        stepRule.step.tmsUpload(
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: jenkinsUtilsStub,
            transportManagementService: tmsStub,
            mtaPath: 'dummy.mtar',
            nodeName: 'myNode',
            credentialsId: 'TMS_ServiceKey',
            nodeExtDescriptorMapping: nodeExtDescriptorMap,
            mtaVersion: '0.0.1',
            useGoStep: false
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
                if(nodeId==1) {
                    calledTmsMethodsWithArgs << "uploadMtaExtDescriptorToNode('${url}', '${token}', ${nodeId}, '${file}', '${mtaVersion}', '${description}', '${namedUser}')"
                    return [id: 123, mtaExtId: "com.sap.piper.tms.test.extension"]
                }
                if(nodeId==2) {
                    calledTmsMethodsWithArgs << "uploadMtaExtDescriptorToNode('${url}', '${token}', ${nodeId}, '${file}', '${mtaVersion}', '${description}', '${namedUser}')"
                    return [id: 456, mtaExtId: "com.sap.piper.tms.test.another.extension"]
                }
            }

            def getNodes(String url, String token) {
                calledTmsMethodsWithArgs << "getNodes('${url}', '${token}')"
                return [nodes: [[id: 1, name: "testNode1"], [id: 2, name: "testNode2"]]]
            }

            def updateMtaExtDescriptor(String url, String token, Long nodeId, Long idOfMtaDescriptor, String file, String mtaVersion, String description, String namedUser) {
                calledTmsMethodsWithArgs << "updateMtaExtDescriptor('${url}', '${token}', ${nodeId}, ${idOfMtaDescriptor}, '${file}', '${mtaVersion}', '${description}', '${namedUser}')"
                return [id: 456, mtaExtId: "com.sap.piper.tms.test.another.extension"]
            }

            def getMtaExtDescriptor(String url, String token, Long nodeId, String mtaId, String mtaVersion) {
                if(mtaVersion=="0.0.1") {
                    calledTmsMethodsWithArgs << "getMtaExtDescriptor('${url}', '${token}', ${nodeId}, '${mtaId}', '${mtaVersion}')"
                    return [:]
                } else {
                    calledTmsMethodsWithArgs << "getMtaExtDescriptor('${url}', '${token}', ${nodeId}, '${mtaId}', '${mtaVersion}')"
                    return ["id": 2, "mtaId": "com.sap.piper.tms.test", "mtaExtId": "com.sap.piper.tms.test.extension", "mtaVersion": "1.2.3"]
                }
            }
        }
    }
}
