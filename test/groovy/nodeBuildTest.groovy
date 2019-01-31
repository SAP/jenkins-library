import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsDockerExecuteRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.Rules

import static org.junit.Assert.assertEquals

class nodeBuildTest extends BasePiperTest {

    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsReadYamlRule yamlRule = new JenkinsReadYamlRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(yamlRule)
        .around(dockerExecuteRule)
        .around(shellRule)
        .around(stepRule)

    @Before
    void init() {
        helper.registerAllowedMethod('fileExists', [String], { s -> s == 'package.json' })
    }

    @Test
    void testNodeBuild() throws Exception {
        stepRule.step.nodeBuild(script: nullScript, dockerImage: 'node:latest')
        assertEquals('node:latest', dockerExecuteRule.dockerParams.dockerImage)
    }
}
