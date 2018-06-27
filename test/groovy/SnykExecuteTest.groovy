import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import org.junit.rules.ExpectedException
import static org.junit.Assert.assertThat

import util.BasePiperTest
import util.JenkinsDockerExecuteRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsLoggingRule
import util.Rules

class SnykExecuteTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsDockerExecuteRule jder = new JenkinsDockerExecuteRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(jder)
        .around(jscr)
        .around(jlr)
        .around(jsr)

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
            if (map.glob == '**/pom.xml')
                return [new File('some-service/pom.xml'), new File('some-other-service/pom.xml')].toArray()
            if (map.glob == '**/package.json')
                return [new File('some-ui/package.json'), new File('some-service-broker/package.json')].toArray()
            return [].toArray()
        })
        helper.registerAllowedMethod('archiveArtifacts', [String], {
            s -> archiveStepPatterns.push(s.toString())
        })
    }

    @Test
    void testUnsupportedScanType() throws Exception {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[ERROR][snykExecute] ScanType \'seagul\' not supported!')

        jsr.step.snykExecute(
            script: nullScript,
            juStabUtils: utils,
            scanType: 'seagul'
        )
    }

    @Test
    void testDefaultsSettings() throws Exception {
        jsr.step.snykExecute(
            script: nullScript,
            juStabUtils: utils
        )

        assertThat(withCredentialsParameters.credentialsId, is('myPassword'))
        assertThat(jder.dockerParams, hasEntry('dockerImage', 'node:8.11.2-stretch'))
        assertThat(jder.dockerParams.stashContent, hasItem('buildDescriptor'))
        assertThat(jder.dockerParams.stashContent, hasItem('opensourceConfiguration'))
    }

    @Test
    void testScanTypeNpm() throws Exception {
        jsr.step.snykExecute(
            script: nullScript,
            juStabUtils: utils
        )
        // asserts
        assertThat(jscr.shell, hasItem('npm install snyk --global --quiet'))
        assertThat(jscr.shell, hasItem('cd \'./\' && npm install --quiet'))
        assertThat(jscr.shell, hasItem('cd \'./\' && snyk monitor && snyk test'))
    }

    @Test
    void testScanTypeNpmWithOrgAndJsonReport() throws Exception {
        jsr.step.snykExecute(
            script: nullScript,
            juStabUtils: utils,
            snykOrg: 'myOrg',
            toJson: true
        )
        // asserts
        assertThat(jscr.shell, hasItem('cd \'./\' && snyk monitor --org=myOrg && snyk test --json > snyk.json'))
        assertThat(archiveStepPatterns, hasItem('snyk.json'))
    }

    @Test
    void testScanTypeMta() throws Exception {
        jsr.step.snykExecute(
            script: nullScript,
            juStabUtils: utils,
            scanType: 'mta'
        )
        // asserts
        assertThat(jscr.shell, hasItem('cd \'some-ui/\' && snyk monitor && snyk test'))
        assertThat(jscr.shell, hasItem('cd \'some-service-broker/\' && snyk monitor && snyk test'))
    }
}
