import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertTrue

class DockerExecuteOnKubernetesTest extends BasePiperTest {
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

    int whichDockerReturnValue = 0
    def bodyExecuted
    def dockerImage
    def containersMap
    def dockerEnvVars
    def dockerWorkspace
    def containerName = ''


    @Before
    void init() {
        bodyExecuted = false
        binding.setVariable('Jenkins', [instance: [pluginManager: [plugins: [new PluginMock('kubernetes'), new PluginMock('docker-workflow')]]]])
        helper.registerAllowedMethod('sh', [Map.class], {return whichDockerReturnValue})
        helper.registerAllowedMethod('container', [Map.class, Closure.class], { Map config, Closure body ->
            containerName = config.name
            body()
        })
        helper.registerAllowedMethod('runInsidePod', [Map.class, Closure.class], { Map config, Closure body ->
            config.containersMap.each { k, v -> dockerImage = k }
            containersMap = config.containersMap
            dockerEnvVars = config.dockerEnvVars
            dockerWorkspace = config.dockerWorkspace
            body()
        })

    }

    @Test
    void testRunOnPod() throws Exception {
        jsr.step.call(script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerOptions: '-it',
            dockerVolumeBind: ['my_vol': '/my_vol'],
            dockerEnvVars: ['http_proxy': 'http://proxy:8000'], dockerWorkspace: '/home/piper') {
            bodyExecuted = true
        }
        assertEquals('container-exec', containerName)
        assertEquals('maven:3.5-jdk-8-alpine', dockerImage)
        assertEquals(['http_proxy': 'http://proxy:8000'], dockerEnvVars)
        assertEquals('/home/piper', dockerWorkspace)
        assertTrue(bodyExecuted)
    }

    @Test
    void testRunOnPodNoDockerImage() throws Exception {
        boolean failed = false
        try {
            jsr.step.call(script: nullScript,
                dockerOptions: '-it',
                dockerEnvVars: ['http_proxy': 'http://proxy:8000']) {
                bodyExecuted = true
            }
        } catch (e) {
            failed = true
        }
        assertTrue(failed)
    }
}
