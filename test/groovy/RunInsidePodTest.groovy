import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.junit.Assert.*

class RunInsidePodTest extends BasePiperTest {
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

    def bodyExecuted
    def podName = ''
    def podLabel = ''
    def containersList = []
    def imageList = []
    def containerName = ''
    def envList = []


    @Before
    void init() {
        containersList = []
        imageList = []
        envList = []
        bodyExecuted = false
        helper.registerAllowedMethod('podTemplate', [Map.class, Closure.class], { Map options, Closure body ->
            podName = options.name
            podLabel = options.label
            options.containers.each { option ->
                containersList.add(option.name)
                imageList.add(option.image)
                envList.add(option.envVars)
            }
            body()
        })
        helper.registerAllowedMethod('node', [String.class, Closure.class], { String nodeName, Closure body -> body()
        })
        helper.registerAllowedMethod('envVar', [Map.class], { Map option -> return option
        })
        helper.registerAllowedMethod('containerTemplate', [Map.class], { Map option -> return option
        })

    }

    @Test
    void testRunInsidePod() throws Exception {
        jsr.step.call(script: nullScript,
            containersMap: ['maven:3.5-jdk-8-alpine': 'mavenexecute']) {
            container(name: 'mavenexecute') {
                bodyExecuted = true
            }
        }
        assertEquals('mavenexecute', containerName)
        assertTrue(containersList.contains('mavenexecute'))
        assertTrue(imageList.contains('maven:3.5-jdk-8-alpine'))
        assertTrue(containersList.contains('jnlp'))
        assertTrue(imageList.contains('jenkinsci/jnlp-slave:latest'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testRunInsidePodWithCustomJnlp() throws Exception {
        nullScript.commonPipelineEnvironment.configuration = ['general': ['jenkinsKubernetes': ['jnlpAgent': 'myJnalpAgent']]]
        jsr.step.call(script: nullScript,
            containersMap: ['maven:3.5-jdk-8-alpine': 'mavenexecute']) {
            container(name: 'mavenexecute') {
                bodyExecuted = true
            }
        }
        assertEquals('mavenexecute', containerName)
        assertTrue(containersList.contains('mavenexecute'))
        assertTrue(imageList.contains('maven:3.5-jdk-8-alpine'))
        assertTrue(containersList.contains('jnlp'))
        assertTrue(imageList.contains('myJnalpAgent'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testRunInsidePodWithCustomWorkspace() throws Exception {
        jsr.step.call(script: nullScript,
            containersMap: ['maven:3.5-jdk-8-alpine': 'mavenexecute'],
            dockerWorkspace: '/home/piper') {
            container(name: 'mavenexecute') {
                bodyExecuted = true
            }
        }
        assertTrue(envList.toString().contains('/home/piper'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testRunInsidePodWithCustomEnv() throws Exception {
        jsr.step.call(script: nullScript,
            containersMap: ['maven:3.5-jdk-8-alpine': 'mavenexecute'],
            dockerEnvVars: ['customEnvKey': 'customEnvValue']) {
            container(name: 'mavenexecute') {
                bodyExecuted = true
            }
        }
        assertTrue(envList.toString().contains('customEnvKey') && envList.toString().contains('customEnvValue'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testRunInsidePodOnlyJnlp() throws Exception {
        jsr.step.call(script: nullScript,
            containersMap: [:],
            dockerEnvVars: ['customEnvKey': 'customEnvValue']) {
            container(name: 'jnlp') {
                bodyExecuted = true
            }
        }
        assertEquals('jnlp', containerName)
        assertTrue(containersList.contains('jnlp'))
        assertTrue(imageList.contains('jenkinsci/jnlp-slave:latest'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testRunOnPodNoContainerMap() throws Exception {
        boolean failed = false
        try {
            jsr.step.call(script: nullScript,
                dockerEnvVars: ['customEnvKey': 'customEnvValue']) {
                container(name: 'mavenexecute') {
                    bodyExecuted = true
                }
            }
        } catch (e) {
            failed = true
        }
        assertTrue(failed)
        assertFalse(bodyExecuted)
    }

    private void container(options, body) {
        containerName = options.name
        body()
    }
}
