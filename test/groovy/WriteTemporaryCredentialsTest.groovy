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
    }

    @Test
    void noCredentials() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[writeTemporaryCredentials] The execution failed, since no credentials are defined. Please provide credentials as a list of maps.')

        stepRule.step.writeTemporaryCredentials(
            script: nullScript,
            stageName: "myStage",
        ){
            bodyExecuted = true
        }
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
    }

    @Test
    void noCredentialsDirectory() {
        def credential = [alias: 'ERP', credentialId: 'erp-credentials']

        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            credentials: [credential]
        ]]]

        thrown.expect(hudson.AbortException)
        thrown.expectMessage("[writeTemporaryCredentials] The execution failed, since no credentialsDirectory is defined. Please provide the path for the credentials file.")

        stepRule.step.writeTemporaryCredentials(
            script: nullScript,
            stageName: "myStage",
        ){
            bodyExecuted = true
        }
    }

    @Test
    void credentialsFileWritten() {
        def credential = [alias: 'ERP', credentialId: 'erp-credentials']
        fileExistsRule.registerExistingFile('systems.yml')

        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            credentials: [credential],
            credentialsDirectory: './',
        ]]]

        stepRule.step.writeTemporaryCredentials(
            script: nullScript,
            stageName: "myStage",
        ){
            bodyExecuted = true
        }

        assertThat(writeFileRule.files['credentials.json'], containsString('"alias":"ERP","username":"test_user","password":"********"'))
        assertThat(shellRule.shell, hasItem('rm -f credentials.json'))
    }
}
