import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is

import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import org.junit.rules.ExpectedException
import static org.junit.Assert.assertThat

import util.BasePiperTest
import util.JenkinsDockerExecuteRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsLoggingRule
import util.Rules

import com.sap.piper.Utils
import com.sap.piper.MapUtils

class SnykExecuteTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(dockerExecuteRule)
        .around(shellRule)
        .around(loggingRule)
        .around(stepRule)

    def withCredentialsParameters
    List archiveStepPatterns

    @Before
    void init() {
        archiveStepPatterns = []
        nullScript.commonPipelineEnvironment.configuration = [
            steps: [
                snykExecute: [
                    snykCredentialsId: 'myPassword'
                ]
            ]
        ]
        helper.registerAllowedMethod('string', [Map], { m -> withCredentialsParameters = m
            return m })
        helper.registerAllowedMethod('withCredentials', [List, Closure], { l, c ->
            binding.setProperty('token', 'test_snyk')
            try {
                c()
            } finally {
                binding.setProperty('token', null)
            }
        })
        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            if (map.glob == "**${File.separator}pom.xml")
                return [new File("some-service${File.separator}pom.xml"), new File("some-other-service${File.separator}pom.xml")].toArray()
            if (map.glob == "**${File.separator}package.json")
                return [new File("some-ui${File.separator}package.json"), new File("some-service-broker${File.separator}package.json")].toArray()
            return [].toArray()
        })
        helper.registerAllowedMethod('archiveArtifacts', [String], {
            s -> archiveStepPatterns.push(s.toString())
        })

        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
    }

    @Test
    void testUnsupportedScanType() throws Exception {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[ERROR][snykExecute] ScanType \'seagul\' not supported!')

        stepRule.step.snykExecute(
            script: nullScript,
            juStabUtils: utils,
            scanType: 'seagul'
        )
    }

    @Test
    void testDefaultsSettings() throws Exception {
        stepRule.step.snykExecute(
            script: nullScript,
            juStabUtils: utils
        )

        assertThat(withCredentialsParameters.credentialsId, is('myPassword'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'node:lts-bookworm'))
        assertThat(dockerExecuteRule.dockerParams.stashContent, hasItem('buildDescriptor'))
        assertThat(dockerExecuteRule.dockerParams.stashContent, hasItem('opensourceConfiguration'))
    }

    @Test
    void testDockerFromCustomStepConfiguration() {

        def expectedImage = 'image:test'
        def expectedEnvVars = ['SNYK_TOKEN':'', 'env1': 'value1', 'env2': 'value2']
        def expectedOptions = '--opt1=val1 --opt2=val2 --opt3'
        def expectedWorkspace = '/path/to/workspace'


        nullScript.commonPipelineEnvironment.configuration = MapUtils.merge(
            nullScript.commonPipelineEnvironment.configuration,
            [steps:[snykExecute:[
                dockerImage: expectedImage,
                dockerOptions: expectedOptions,
                dockerEnvVars: expectedEnvVars,
                dockerWorkspace: expectedWorkspace
            ]]])

        stepRule.step.snykExecute(
            script: nullScript,
            juStabUtils: utils
        )

        assert expectedImage == dockerExecuteRule.dockerParams.dockerImage
        assert expectedOptions == dockerExecuteRule.dockerParams.dockerOptions
        assert expectedEnvVars.equals(dockerExecuteRule.dockerParams.dockerEnvVars)
        assert expectedWorkspace == dockerExecuteRule.dockerParams.dockerWorkspace
    }

    @Test
    void testScanTypeNpm() throws Exception {
        stepRule.step.snykExecute(
            script: nullScript,
            juStabUtils: utils
        )
        // asserts
        assertThat(shellRule.shell, hasItem('npm install snyk --global --quiet'))
        assertThat(shellRule.shell, hasItem('cd \'./\' && npm install --quiet'))
        assertThat(shellRule.shell, hasItem('cd \'./\' && snyk monitor && snyk test'))
    }

    @Test
    void testScanTypeNpmWithOrgAndJsonReport() throws Exception {
        stepRule.step.snykExecute(
            script: nullScript,
            juStabUtils: utils,
            snykOrg: 'myOrg',
            toJson: true
        )
        // asserts
        assertThat(shellRule.shell, hasItem("cd './' && snyk monitor --org=myOrg && snyk test --json > snyk.json".toString()))
        assertThat(archiveStepPatterns, hasItem('snyk.json'))
    }

    @Test
    void testScanTypeMta() throws Exception {
        stepRule.step.snykExecute(
            script: nullScript,
            juStabUtils: utils,
            scanType: 'mta'
        )
        // asserts
        assertThat(shellRule.shell, hasItem("cd 'some-ui${File.separator}' && snyk monitor && snyk test".toString()))
        assertThat(shellRule.shell, hasItem("cd 'some-service-broker${File.separator}' && snyk monitor && snyk test".toString()))
    }
}
