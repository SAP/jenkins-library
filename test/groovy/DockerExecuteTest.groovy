import com.sap.piper.k8s.ContainerMap
import com.sap.piper.JenkinsUtils
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsStepRule
import util.PluginMock
import util.Rules

import static org.junit.Assert.assertTrue
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertFalse

class DockerExecuteTest extends BasePiperTest {
    private DockerMock docker
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(jlr)
        .around(jsr)

    int whichDockerReturnValue = 0
    def bodyExecuted
    def containerName

    @Before
    void init() {
        bodyExecuted = false
        docker = new DockerMock()
        JenkinsUtils.metaClass.static.isPluginActive = {def s -> new PluginMock(s).isActive()}
        binding.setVariable('docker', docker)
        helper.registerAllowedMethod('sh', [Map.class], {return whichDockerReturnValue})
    }

    @Test
    void testExecuteInsideContainerOfExistingPod() throws Exception {
        helper.registerAllowedMethod('container', [String.class, Closure.class], { String container, Closure body ->
            containerName = container
            body()
        })
        binding.setVariable('env', [POD_NAME: 'testpod', ON_K8S: 'true'])
        ContainerMap.instance.setMap(['testpod': ['maven:3.5-jdk-8-alpine': 'mavenexec']])
        jsr.step.call(script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerEnvVars: ['http_proxy': 'http://proxy:8000']) {
            bodyExecuted = true
        }
        assertTrue(jlr.log.contains('Executing inside a Kubernetes Container'))
        assertEquals('mavenexec', containerName)
        assertTrue(bodyExecuted)
     }

    @Test
    void testExecuteInsideNewlyCreatedPod() throws Exception {
        helper.registerAllowedMethod('dockerExecuteOnKubernetes', [Map.class, Closure.class], { Map config, Closure body -> body() })
        binding.setVariable('env', [ON_K8S: 'true'])
        ContainerMap.instance.setMap(['testpod': ['maven:3.5-jdk-8-alpine': 'mavenexec']])
        jsr.step.call(script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerEnvVars: ['http_proxy': 'http://proxy:8000']) {
            bodyExecuted = true
        }
        assertTrue(jlr.log.contains('Executing inside a Kubernetes Pod'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testExecuteInsidePodWithEmptyContainerMap() throws Exception {
        helper.registerAllowedMethod('dockerExecuteOnKubernetes', [Map.class, Closure.class], { Map config, Closure body -> body() })
        binding.setVariable('env', [POD_NAME: 'testpod', ON_K8S: 'true'])
        ContainerMap.instance.setMap([:])
        jsr.step.call(script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerEnvVars: ['http_proxy': 'http://proxy:8000']) {
            bodyExecuted = true
        }
        assertTrue(jlr.log.contains('Executing inside a Kubernetes Pod'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testExecuteInsidePodWithStageKeyEmptyValue() throws Exception {
        helper.registerAllowedMethod('dockerExecuteOnKubernetes', [Map.class, Closure.class], { Map config, Closure body -> body() })
        binding.setVariable('env', [POD_NAME: 'testpod', ON_K8S: 'true'])
        ContainerMap.instance.setMap(['testpod':[:]])
        jsr.step.call(script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerEnvVars: ['http_proxy': 'http://proxy:8000']) {
            bodyExecuted = true
        }
        assertTrue(jlr.log.contains('Executing inside a Kubernetes Pod'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testExecuteInsideDockerContainer() throws Exception {
        jsr.step.call(script: nullScript, dockerImage: 'maven:3.5-jdk-8-alpine') {
            bodyExecuted = true
        }
        assertEquals('maven:3.5-jdk-8-alpine', docker.getImageName())
        assertTrue(docker.isImagePulled())
        assertEquals('--env http_proxy --env https_proxy --env no_proxy --env HTTP_PROXY --env HTTPS_PROXY --env NO_PROXY', docker.getParameters().trim())
        assertTrue(bodyExecuted)
    }

    @Test
    void testExecuteInsideDockerNoScript() throws Exception {
        jsr.step.call(dockerImage: 'maven:3.5-jdk-8-alpine') {
            bodyExecuted = true
        }
        assertEquals('maven:3.5-jdk-8-alpine', docker.getImageName())
        assertTrue(docker.isImagePulled())
        assertEquals('--env http_proxy --env https_proxy --env no_proxy --env HTTP_PROXY --env HTTPS_PROXY --env NO_PROXY', docker.getParameters().trim())
        assertTrue(bodyExecuted)
    }

    @Test
    void testExecuteInsideDockerContainerWithParameters() throws Exception {
        jsr.step.call(script: nullScript,
                      dockerImage: 'maven:3.5-jdk-8-alpine',
                      dockerOptions: '-it',
                      dockerVolumeBind: ['my_vol': '/my_vol'],
                      dockerEnvVars: ['http_proxy': 'http://proxy:8000']) {
            bodyExecuted = true
        }
        assertTrue(docker.getParameters().contains('--env https_proxy '))
        assertTrue(docker.getParameters().contains('--env http_proxy=http://proxy:8000'))
        assertTrue(docker.getParameters().contains('-it'))
        assertTrue(docker.getParameters().contains('--volume my_vol:/my_vol'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testExecuteInsideDockerContainerWithDockerOptionsList() throws Exception {
        jsr.step.call(script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerOptions: ['-it', '--network=my-network'],
            dockerEnvVars: ['http_proxy': 'http://proxy:8000']) {
            bodyExecuted = true
        }
        assertTrue(docker.getParameters().contains('--env http_proxy=http://proxy:8000'))
        assertTrue(docker.getParameters().contains('-it'))
        assertTrue(docker.getParameters().contains('--network=my-network'))
    }

    @Test
    void testDockerNotInstalledResultsInLocalExecution() throws Exception {
        whichDockerReturnValue = 1
        jsr.step.call(script: nullScript,
            dockerOptions: '-it') {
            bodyExecuted = true
        }
        assertTrue(jlr.log.contains('No docker environment found'))
        assertTrue(jlr.log.contains('Running on local environment'))
        assertTrue(bodyExecuted)
        assertFalse(docker.isImagePulled())
    }

    private class DockerMock {
        private String imageName
        private boolean imagePulled = false
        private String parameters

        DockerMock image(String imageName) {
            this.imageName = imageName
            return this
        }

        void pull() {
            imagePulled = true
        }

        void inside(String parameters, body) {
            this.parameters = parameters
            body()
        }

        String getImageName() {
            return imageName
        }

        boolean isImagePulled() {
            return imagePulled
        }

        String getParameters() {
            return parameters
        }
    }
}
