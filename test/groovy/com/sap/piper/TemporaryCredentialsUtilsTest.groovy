package com.sap.piper

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsFileExistsRule
import util.JenkinsLoggingRule
import util.JenkinsReadFileRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.junit.Assert.assertThat
import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertTrue

class TemporaryCredentialsUtilsTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsCredentialsRule credentialsRule = new JenkinsCredentialsRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
    private JenkinsFileExistsRule fileExistsRule = new JenkinsFileExistsRule(this, [])
    private JenkinsReadFileRule readFileRule = new JenkinsReadFileRule(this, null)
    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)

    def bodyExecuted
    TemporaryCredentialsUtils credUtils

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(readYamlRule)
        .around(credentialsRule)
        .around(loggingRule)
        .around(writeFileRule)
        .around(fileExistsRule)
        .around(readFileRule)
        .around(shellRule)

    @Before
    void init() {
        bodyExecuted = false

        credentialsRule.reset()
            .withCredentials('erp-credentials', 'test_user', '********')
            .withCredentials('testCred2', 'test_other', '**')

        credUtils = new TemporaryCredentialsUtils(nullScript)
    }

    @Test
    void singleCredentialsFileWritten() {
        def credential = [alias: 'ERP', credentialId: 'erp-credentials']
        def directories = ['./', 'integration-tests/src/test/resources']
        def filename = 'credentials.json'
        fileExistsRule.registerExistingFile('./systems.yml')

        credUtils.writeCredentials([credential], directories, filename )

        assertThat(writeFileRule.files['./credentials.json'], containsString('"alias":"ERP","username":"test_user","password":"********"'))
    }

    @Test
    void twoCredentialsFilesWritten() {
        def credential = [alias: 'ERP', credentialId: 'erp-credentials']
        def directories = ['./', 'integration-tests/src/test/resources']
        def filename = 'credentials.json'
        fileExistsRule.registerExistingFile('./systems.yml')
        fileExistsRule.registerExistingFile('integration-tests/src/test/resources/systems.yml')

        credUtils.writeCredentials([credential], directories, filename )

        assertThat(writeFileRule.files["./credentials.json"], containsString('"alias":"ERP","username":"test_user","password":"********"'))
        assertThat(writeFileRule.files["integration-tests/src/test/resources/credentials.json"], containsString('"alias":"ERP","username":"test_user","password":"********"'))
    }

    @Test
    void credentialsFileNotWrittenWithEmptyList() {
        def directories = ['./', 'integration-tests/src/test/resources']
        def filename = 'credentials.json'
        fileExistsRule.registerExistingFile('systems.yml')

        credUtils.writeCredentials([], directories, filename )

        loggingRule.expect('Not writing any credentials.')
    }

    @Test
    void systemsFileNotExists() {
        def credential = [alias: 'ERP', credentialId: 'erp-credentials']
        def directories = ['./', 'integration-tests/src/test/resources']
        def filename = 'credentials.json'
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("None of the directories [./, integration-tests/src/test/resources/] contains any of the files systems.yml, systems.yaml or systems.json. " +
            "One of those files is required in order to activate the integration test credentials configured in the pipeline configuration file of this project. " +
            "Please add the file as explained in project 'Piper' documentation.")

        credUtils.writeCredentials([credential], directories, filename )
    }

    @Test
    void credentialsFileDeleted() {
        def directories = ['./', 'integration-tests/src/test/resources']
        def filename = 'credentials.json'
        fileExistsRule.registerExistingFile('systems.yml')
        fileExistsRule.registerExistingFile('./credentials.json')

        credUtils.deleteCredentials(directories, filename )

        assertThat(shellRule.shell, hasItem('rm -f ./credentials.json'))
    }

    @Test
    void handleTemporaryCredentials() {
        def credential = [alias: 'ERP', credentialId: 'erp-credentials']
        def directories = ['./', 'integration-tests/src/test/resources']
        fileExistsRule.registerExistingFile('./systems.yml')
        fileExistsRule.registerExistingFile('./credentials.json')

        credUtils.handleTemporaryCredentials([credential], directories) {
            bodyExecuted = true
        }
        assertTrue(bodyExecuted)
        assertThat(writeFileRule.files['./credentials.json'], containsString('"alias":"ERP","username":"test_user","password":"********"'))
        assertThat(shellRule.shell, hasItem('rm -f ./credentials.json'))
    }

    @Test
    void handleTemporaryCredentialsNoDirectories() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("This should not happen: Directories for credentials files not specified.")

        credUtils.handleTemporaryCredentials([], []){
            bodyExecuted = true
        }
    }

    @Test
    void handleTemporaryCredentialsNoCredentials() {
        def directories = ['./', 'integration-tests/src/test/resources']
        credUtils.handleTemporaryCredentials([], directories){
            bodyExecuted = true
        }
        assertTrue(bodyExecuted)
        assertThat(writeFileRule.files.keySet(), hasSize(0))
        assertThat(shellRule.shell, hasSize(0))
    }
}
