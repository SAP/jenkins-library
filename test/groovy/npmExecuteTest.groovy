import static org.junit.Assert.assertEquals
import hudson.AbortException
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsDockerExecuteRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.Rules

class npmExecuteTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException().none()
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsReadYamlRule yamlRule = new JenkinsReadYamlRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(yamlRule)
        .around(dockerExecuteRule)
        .around(shellRule)
        .around(stepRule)

    @Before
    void init() {
        helper.registerAllowedMethod 'fileExists', [String], { s -> s == 'package.json' }
    }

    @Test
    void testNpmExecute() {
        stepRule.step.npmExecute(script: nullScript, dockerImage: 'node:8-stretch', npmCommand: 'run build') {}
        assertEquals 'node:8-stretch', dockerExecuteRule.dockerParams.dockerImage
    }

    @Test
    void testNoPackageJson() {
        helper.registerAllowedMethod 'fileExists', [String], { false }
        thrown.expect AbortException
        thrown.expectMessage '[npmExecute] package.json is not found.'
        stepRule.step.npmExecute(script: nullScript, dockerImage: 'node:8-stretch', npmCommand: 'run build') {}
    }
}
