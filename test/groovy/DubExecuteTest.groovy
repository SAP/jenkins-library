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

class DubExecuteTest extends BasePiperTest {

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
        helper.registerAllowedMethod 'fileExists', [String], { s -> s == 'dub.json' }
    }

    @Test
    void testDubExecute() {
        stepRule.step.dubExecute(script: nullScript, dockerImage: 'dlang2/dmd-ubuntu:latest')
        assertEquals 'dlang2/dmd-ubuntu:latest', dockerExecuteRule.dockerParams.dockerImage
    }

    @Test
    void testDubExecuteWithClosure() {
        stepRule.step.dubExecute(script: nullScript, dockerImage: 'dlang2/dmd-ubuntu:latest', dubCommand: 'build') { }
        assert shellRule.shell.find { c -> c.contains('dub build') }
    }

    @Test
    void testNoDubJson() {
        helper.registerAllowedMethod 'fileExists', [String], { false }
        thrown.expect AbortException
        thrown.expectMessage '[dubExecute] Neither dub.json nor dub.sdl was found.'
        stepRule.step.dubExecute(script: nullScript, dockerImage: 'dlang2/dmd-ubuntu:latest', dubCommand: 'build')
    }
}
