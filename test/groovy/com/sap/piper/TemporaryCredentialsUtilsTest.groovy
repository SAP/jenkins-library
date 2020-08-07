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
import util.JenkinsStepRule
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
    void credentialsFileWritten() {
        def credential = [alias: 'ERP', credentialId: 'erp-credentials']
        def directory = './'
        def filename = 'credentials.json'
        fileExistsRule.registerExistingFile('systems.yml')

        credUtils.writeCredentials([credential], directory, filename )

        assertThat(writeFileRule.files['credentials.json'], containsString('"alias":"ERP","username":"test_user","password":"********"'))
    }

    @Test
    void credentialsFileNotWrittenWithEmptyList() {
        def directory = './'
        def filename = 'credentials.json'
        fileExistsRule.registerExistingFile('systems.yml')

        credUtils.writeCredentials([], directory, filename )

        loggingRule.expect('Not writing any credentials.')
    }

    @Test
    void credentialsFileDeleted() {
        def directory = './'
        def filename = 'credentials.json'
        fileExistsRule.registerExistingFile('systems.yml')

        credUtils.deleteCredentials(directory, filename )

        assertThat(shellRule.shell, hasItem('rm -f credentials.json'))
    }

    @Test
    void systemsFileNotExists() {
        def directory = './'
        thrown.expect(hudson.AbortException)
        thrown.expectMessage("The directory ${directory} does not contain any of the files systems.yml, systems.yaml or systems.json. " +
            "One of those files is required in order to activate the integration test credentials configured in the pipeline configuration file of this project. " +
            "Please add the file as explained in the SAP Cloud SDK documentation.")

        credUtils.assertSystemsFileExists(directory)
    }

    @Test
    void handleTemporaryCredentials() {
        def credential = [alias: 'ERP', credentialId: 'erp-credentials']
        def directory = './'
        fileExistsRule.registerExistingFile('systems.yml')

        credUtils.handleTemporaryCredentials([credential], directory) {
            bodyExecuted = true
        }
        assertTrue(bodyExecuted)
        assertThat(writeFileRule.files['credentials.json'], containsString('"alias":"ERP","username":"test_user","password":"********"'))
        assertThat(shellRule.shell, hasItem('rm -f credentials.json'))
    }
}
