import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsFileExistsRule
import util.JenkinsReadFileRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules
import static org.junit.Assert.assertThat
import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertTrue
import static org.junit.Assert.assertFalse
import com.sap.piper.Utils

class WriteTemporaryCredentialsTest extends BasePiperTest {
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsCredentialsRule credentialsRule = new JenkinsCredentialsRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
    private JenkinsFileExistsRule fileExistsRule = new JenkinsFileExistsRule(this, [])
    private JenkinsReadFileRule readFileRule = new JenkinsReadFileRule(this, null)
    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)

    def bodyExecuted

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(readYamlRule)
        .around(credentialsRule)
        .around(stepRule)
        .around(writeFileRule)
        .around(fileExistsRule)
        .around(readFileRule)
        .around(shellRule)

    @Before
    void init() {
        bodyExecuted = false

        helper.registerAllowedMethod("deleteDir", [], null)

        credentialsRule.reset()
            .withCredentials('erp-credentials', 'test_user', '********')
            .withCredentials('testCred2', 'test_other', '**')

        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
    }

    @Test
    void noCredentials() {
        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            credentialsDirectories: ['./', 'integration-test/'],
        ]]]
        stepRule.step.writeTemporaryCredentials(
            script: nullScript,
            stageName: "myStage",
        ){
            bodyExecuted = true
        }
        assertTrue(bodyExecuted)
        assertThat(writeFileRule.files.keySet(), hasSize(0))
        assertThat(shellRule.shell, hasSize(0))
    }

    @Test
    void credentialsNoList() {
        def credential = "id"

        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            credentials: credential
        ]]]

        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[writeTemporaryCredentials] The execution failed, since credentials is not a list. Please provide credentials as a list of maps.')

        stepRule.step.writeTemporaryCredentials(
            script: nullScript,
            stageName: "myStage",
        ){
            bodyExecuted = true
        }
        assertFalse(bodyExecuted)
    }

    @Test
    void noCredentialsDirectories() {
        def credential = [alias: 'ERP', credentialId: 'erp-credentials']

        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            credentials: [credential],
            credentialsDirectories: []
        ]]]

        thrown.expect(hudson.AbortException)
        thrown.expectMessage("[writeTemporaryCredentials] The execution failed, since no credentialsDirectories are defined. Please provide a list of paths for the credentials files.")

        stepRule.step.writeTemporaryCredentials(
            script: nullScript,
            stageName: "myStage",
        ){
            bodyExecuted = true
        }
        assertFalse(bodyExecuted)
    }

    @Test
    void credentialsDirectoriesNoList() {
        def credential = [alias: 'ERP', credentialId: 'erp-credentials']

        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            credentials: [credential],
            credentialsDirectories: './',
        ]]]

        thrown.expect(hudson.AbortException)
        thrown.expectMessage("[writeTemporaryCredentials] The execution failed, since credentialsDirectories is not a list. Please provide credentialsDirectories as a list of paths.")

        stepRule.step.writeTemporaryCredentials(
            script: nullScript,
            stageName: "myStage",
        ){
            bodyExecuted = true
        }
        assertFalse(bodyExecuted)
    }

    @Test
    void credentialsFileWrittenAndRemoved() {
        def credential = [alias: 'ERP', credentialId: 'erp-credentials']
        fileExistsRule.registerExistingFile('./systems.yml')
        fileExistsRule.registerExistingFile('./credentials.json')

        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            credentials: [credential],
            credentialsDirectories: ['./', 'integration-test/'],
        ]]]

        stepRule.step.writeTemporaryCredentials(
            script: nullScript,
            stageName: "myStage",
        ){
            bodyExecuted = true
        }

        assertTrue(bodyExecuted)
        assertThat(writeFileRule.files['./credentials.json'], containsString('"alias":"ERP","username":"test_user","password":"********"'))
        assertThat(shellRule.shell, hasItem('rm -f ./credentials.json'))
        assertThat(writeFileRule.files.size(), is(1))
    }
}
