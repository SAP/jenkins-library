import static org.junit.Assert.assertEquals
import hudson.AbortException
import org.junit.After
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
import com.sap.piper.Utils

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
        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
    }

    @Test
    void testDubExecute() {
        stepRule.step.dubExecute(script: nullScript, dockerImage: 'dlang2/dmd-ubuntu:latest')
        assertEquals 'dlang2/dmd-ubuntu:latest', dockerExecuteRule.dockerParams.dockerImage
    }

    @Test
    void testDockerFromCustomStepConfiguration() {

        def expectedImage = 'image:test'
        def expectedEnvVars = ['env1': 'value1', 'env2': 'value2']
        def expectedOptions = '--opt1=val1 --opt2=val2 --opt3'
        def expectedWorkspace = '/path/to/workspace'
        
        nullScript.commonPipelineEnvironment.configuration = [steps:[dubExecute:[
            dockerImage: expectedImage, 
            dockerOptions: expectedOptions,
            dockerEnvVars: expectedEnvVars,
            dockerWorkspace: expectedWorkspace
            ]]]

        stepRule.step.dubExecute(
            script: nullScript,
            juStabUtils: utils
        )
        
        assert expectedImage == dockerExecuteRule.dockerParams.dockerImage
        assert expectedOptions == dockerExecuteRule.dockerParams.dockerOptions
        assert expectedEnvVars.equals(dockerExecuteRule.dockerParams.dockerEnvVars)
        assert expectedWorkspace == dockerExecuteRule.dockerParams.dockerWorkspace
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
