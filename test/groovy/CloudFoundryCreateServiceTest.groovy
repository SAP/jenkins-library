import com.sap.piper.JenkinsUtils
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsEnvironmentRule
import util.JenkinsDockerExecuteRule
import util.JenkinsFileExistsRule
import util.JenkinsLoggingRule
import util.JenkinsReadFileRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.JenkinsReadYamlRule
import util.Rules

import static org.hamcrest.Matchers.stringContainsInOrder
import static org.junit.Assert.*

import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.containsString

class CloudFoundryCreateServiceTest extends BasePiperTest {

    private File tmpDir = File.createTempDir()
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsEnvironmentRule environmentRule = new JenkinsEnvironmentRule(this)
    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this)
    private JenkinsFileExistsRule fileExistsRule = new JenkinsFileExistsRule(this)
    private JenkinsCredentialsRule credentialsRule = new JenkinsCredentialsRule(this).withCredentials('test_cfCredentialsId', 'test_cf', '********')

    private writeInfluxMap = [:]

    class JenkinsUtilsMock extends JenkinsUtils {
        def isJobStartedByUser() {
            return true
        }
    }

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(readYamlRule)
        .around(thrown)
        .around(loggingRule)
        .around(shellRule)
        .around(dockerExecuteRule)
        .around(environmentRule)
        .around(fileExistsRule)
        .around(credentialsRule)
        .around(stepRule) // needs to be activated after dockerExecuteRule, otherwise executeDocker is not mocked

    @Before
    void init() {
        helper.registerAllowedMethod('influxWriteData', [Map.class], {m ->
            writeInfluxMap = m
        })
        fileExistsRule.registerExistingFile('test.yml')
    }

    @Test
    void testVarsListNotAList() {               
        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[cloudFoundryCreateService] ERROR: Parameter config.cloudFoundry.manifestVariables is not a List!')

        stepRule.step.cloudFoundryCreateService([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',           
            cfServiceManifest: 'test.yml',
            cfManifestVariables: 'notAList'
        ])       
    }

    @Test
    void testVarsListEntryIsNotAMap() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[cloudFoundryCreateService] ERROR: Parameter config.cloudFoundry.manifestVariables.notAMap is not a Map!')

        stepRule.step.cloudFoundryCreateService([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',           
            cfServiceManifest: 'test.yml',
            cfManifestVariables: ['notAMap']
        ])       
    }

    @Test
    void testVarsFilesListIsNotAList() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[cloudFoundryCreateService] ERROR: Parameter config.cloudFoundry.manifestVariablesFiles is not a List!')

        stepRule.step.cloudFoundryCreateService([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',           
            cfServiceManifest: 'test.yml',
            cfManifestVariablesFiles: 'notAList'
        ])       
    }

    @Test
    void testRunCreateServicePushPlugin() {
        stepRule.step.cloudFoundryCreateService([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',           
            cfServiceManifest: 'test.yml'
        ]) 

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))        
        assertThat(shellRule.shell, hasItem(containsString("cf login -u 'test_cf' -p '********' -a https://api.cf.eu10.hana.ondemand.com -o 'testOrg' -s 'testSpace'")))
        assertThat(shellRule.shell, hasItem(containsString(" cf create-service-push --no-push --service-manifest 'test.yml'")))
        assertThat(shellRule.shell, hasItem(containsString("cf logout")))
    }

    @Test
    void testWithVariableSubstitutionFromVarsListAndVarsFile() {
        String varsFileName='vars.yml'
        fileExistsRule.registerExistingFile(varsFileName)
        List varsList = [["appName" : "testApplicationFromVarsList"]]

        stepRule.step.cloudFoundryCreateService([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfServiceManifest: 'test.yml',
            cfManifestVariablesFiles: [varsFileName],
            cfManifestVariables: varsList
        ])

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))        
        assertThat(shellRule.shell, hasItem(containsString("cf login -u 'test_cf' -p '********' -a https://api.cf.eu10.hana.ondemand.com -o 'testOrg' -s 'testSpace'")))
        assertThat(shellRule.shell, hasItem(containsString("cf create-service-push --no-push --service-manifest 'test.yml' --var appName='testApplicationFromVarsList' --vars-file 'vars.yml'")))
        assertThat(shellRule.shell, hasItem(containsString("cf logout")))
    }

    @Test
    void testEscapesUsernameAndPasswordInShellCall() {
        credentialsRule.credentials.put('escape_cfCredentialsId',[user:"aUserWithA'",passwd:"passHasA'"])

        stepRule.step.cloudFoundryCreateService([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'escape_cfCredentialsId',
            cfServiceManifest: 'test.yml'
        ])

        assertThat(shellRule.shell, hasItem(containsString("""cf login -u 'aUserWithA'"'"'' -p 'passHasA'"'"'' -a https://api.cf.eu10.hana.ondemand.com -o 'testOrg' -s 'testSpace'""")))
    }

    @Test
    void testEscapesSpaceNameInShellCall() {
        stepRule.step.cloudFoundryCreateService([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: "testSpaceWith'",
            cfCredentialsId: 'test_cfCredentialsId',
            cfServiceManifest: 'test.yml'
        ])
        assertThat(shellRule.shell, hasItem(containsString("""cf login -u 'test_cf' -p '********' -a https://api.cf.eu10.hana.ondemand.com -o 'testOrg' -s 'testSpaceWith'"'"''""")))
    }

    @Test
    void testEscapesOrgNameInShellCall() {
        stepRule.step.cloudFoundryCreateService([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: "testOrgWith'",
            cfSpace: "testSpace",
            cfCredentialsId: 'test_cfCredentialsId',
            cfServiceManifest: 'test.yml'
        ])
        assertThat(shellRule.shell, hasItem(containsString("""cf login -u 'test_cf' -p '********' -a https://api.cf.eu10.hana.ondemand.com -o 'testOrgWith'"'"'' -s 'testSpace'""")))
    }

    @Test
    void testWithVariableSubstitutionFromVarsListGetsEscaped() {
        List varsList = [["appName" : "testApplicationFromVarsListWith'"]]

        stepRule.step.cloudFoundryCreateService([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfServiceManifest: 'test.yml',
            cfManifestVariables: varsList
        ])

        assertThat(shellRule.shell, hasItem(containsString("""cf create-service-push --no-push --service-manifest 'test.yml' --var appName='testApplicationFromVarsListWith'"'"''""")))
    }

    @Test
    void testWithVariableSubstitutionFromVarsFilesGetsEscaped() {
        String varsFileName="varsWith'.yml"
        fileExistsRule.registerExistingFile(varsFileName)

        stepRule.step.cloudFoundryCreateService([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfServiceManifest: 'test.yml',
            cfManifestVariablesFiles: [varsFileName]
        ])

        assertThat(shellRule.shell, hasItem(containsString("""cf create-service-push --no-push --service-manifest 'test.yml' --vars-file 'varsWith'"'"'.yml'""")))
    }

    @Test
    void testCfLogoutHappensEvenWhenCreateServiceFails() {

        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[cloudFoundryCreateService] ERROR: The execution of the create-service-push plugin failed, see the logs above for more details.')

        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX,/(create-service-push)/,128)

        stepRule.step.cloudFoundryCreateService([
            script: nullScript,
            juStabUtils: utils,
            jenkinsUtilsStub: new JenkinsUtilsMock(),
            deployTool: 'cf_native',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfCredentialsId: 'test_cfCredentialsId',
            cfServiceManifest: 'test.yml'
        ])

        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))
        assertThat(shellRule.shell, hasItem(containsString("cf login -u 'test_cf' -p '********' -a https://api.cf.eu10.hana.ondemand.com -o 'testOrg' -s 'testSpace'")))
        assertThat(shellRule.shell, hasItem(containsString(" cf create-service-push --no-push --service-manifest 'test.yml'")))
        assertThat(shellRule.shell, hasItem(containsString("cf logout")))
    } 
}
