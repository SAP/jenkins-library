import com.sap.piper.k8s.ContainerMap

import com.sap.piper.JenkinsUtils
import com.sap.piper.SidecarUtils
import com.sap.piper.Utils
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.PluginMock
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertFalse

class DockerExecuteTest extends BasePiperTest {
    private DockerMock docker
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(loggingRule)
        .around(stepRule)
        .around(shellRule)

    def bodyExecuted
    def containerName

    @Before
    void init() {
        bodyExecuted = false
        docker = new DockerMock()
        JenkinsUtils.metaClass.static.isPluginActive = { def s -> new PluginMock(s).isActive() }
        binding.setVariable('docker', docker)
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, "docker .*", 0)
        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
    }

    @Test
    void testExecuteInsideContainerOfExistingPod() throws Exception {
        List usedDockerEnvVars
        helper.registerAllowedMethod('container', [String.class, Closure.class], { String container, Closure body ->
            containerName = container
            body()
        })
        helper.registerAllowedMethod('withEnv', [List.class, Closure.class], { List envVars, Closure body ->
            usedDockerEnvVars = envVars
            body()
        })
        binding.setVariable('env', [POD_NAME: 'testpod', ON_K8S: 'true'])
        ContainerMap.instance.setMap(['testpod': ['maven:3.5-jdk-8-alpine': 'mavenexec']])
        stepRule.step.dockerExecute(script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerEnvVars: ['http_proxy': 'http://proxy:8000']) {
            bodyExecuted = true
        }
        assertTrue(loggingRule.log.contains('Executing inside a Kubernetes Container'))
        assertEquals('mavenexec', containerName)
        assertEquals(usedDockerEnvVars[0].toString(), "http_proxy=http://proxy:8000")
        assertTrue(bodyExecuted)
    }

    @Test
    void testExecuteInsideNewlyCreatedPod() throws Exception {
        helper.registerAllowedMethod('dockerExecuteOnKubernetes', [Map.class, Closure.class], { Map config, Closure body -> body() })
        binding.setVariable('env', [ON_K8S: 'true'])
        ContainerMap.instance.setMap(['testpod': ['maven:3.5-jdk-8-alpine': 'mavenexec']])
        stepRule.step.dockerExecute(script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerEnvVars: ['http_proxy': 'http://proxy:8000']) {
            bodyExecuted = true
        }
        assertTrue(loggingRule.log.contains('Executing inside a Kubernetes Pod'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testExecuteInsidePodWithEmptyContainerMap() throws Exception {
        helper.registerAllowedMethod('dockerExecuteOnKubernetes', [Map.class, Closure.class], { Map config, Closure body -> body() })
        binding.setVariable('env', [POD_NAME: 'testpod', ON_K8S: 'true'])
        ContainerMap.instance.setMap([:])
        stepRule.step.dockerExecute(script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerEnvVars: ['http_proxy': 'http://proxy:8000']) {
            bodyExecuted = true
        }
        assertTrue(loggingRule.log.contains('Executing inside a Kubernetes Pod'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testExecuteInsidePodWithStageKeyEmptyValue() throws Exception {
        helper.registerAllowedMethod('dockerExecuteOnKubernetes', [Map.class, Closure.class], { Map config, Closure body -> body() })
        binding.setVariable('env', [POD_NAME: 'testpod', ON_K8S: 'true'])
        ContainerMap.instance.setMap(['testpod': [:]])
        stepRule.step.dockerExecute(script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerEnvVars: ['http_proxy': 'http://proxy:8000']) {
            bodyExecuted = true
        }
        assertTrue(loggingRule.log.contains('Executing inside a Kubernetes Pod'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testExecuteInsidePodWithCustomCommandAndShell() throws Exception {
        Map kubernetesConfig = [:]
        helper.registerAllowedMethod('dockerExecuteOnKubernetes', [Map.class, Closure.class], { Map config, Closure body ->
            kubernetesConfig = config
            return body()
        })
        binding.setVariable('env', [ON_K8S: 'true'])
        stepRule.step.dockerExecute(
            script: nullScript,
            containerCommand: '/busybox/tail -f /dev/null',
            containerShell: '/busybox/sh',
            dockerImage: 'maven:3.5-jdk-8-alpine'
        ) {
            bodyExecuted = true
        }
        assertTrue(loggingRule.log.contains('Executing inside a Kubernetes Pod'))
        assertThat(kubernetesConfig.containerCommand, is('/busybox/tail -f /dev/null'))
        assertThat(kubernetesConfig.containerShell, is('/busybox/sh'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testExecuteInsidePodWithCustomUserShort() throws Exception {
        Map kubernetesConfig = [:]
        helper.registerAllowedMethod('dockerExecuteOnKubernetes', [Map.class, Closure.class], { Map config, Closure body ->
            kubernetesConfig = config
            return body()
        })
        binding.setVariable('env', [ON_K8S: 'true'])
        stepRule.step.dockerExecute(
            script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerOptions: ["-u 0:0", "-v foo:bar"]
        ) {
            bodyExecuted = true
        }

        assertTrue(loggingRule.log.contains('Executing inside a Kubernetes Pod'))
        assertThat(kubernetesConfig.securityContext, is([
            'runAsUser': 0,
            'runAsGroup': 0
        ]))
        assertTrue(bodyExecuted)
    }

    @Test
    void testExecuteInsidePodWithCustomUserLong() throws Exception {
        Map kubernetesConfig = [:]
        helper.registerAllowedMethod('dockerExecuteOnKubernetes', [Map.class, Closure.class], { Map config, Closure body ->
            kubernetesConfig = config
            return body()
        })
        binding.setVariable('env', [ON_K8S: 'true'])
        stepRule.step.dockerExecute(
            script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerOptions: ["--user 0:0", "-v foo:bar"]
        ) {
            bodyExecuted = true
        }

        assertTrue(loggingRule.log.contains('Executing inside a Kubernetes Pod'))
        assertThat(kubernetesConfig.securityContext, is([
            'runAsUser': 0,
            'runAsGroup': 0
        ]))
        assertTrue(bodyExecuted)
    }

    @Test
    void testExecuteInsidePodWithCustomUserNoGroup() throws Exception {
        Map kubernetesConfig = [:]
        helper.registerAllowedMethod('dockerExecuteOnKubernetes', [Map.class, Closure.class], { Map config, Closure body ->
            kubernetesConfig = config
            return body()
        })
        binding.setVariable('env', [ON_K8S: 'true'])
        stepRule.step.dockerExecute(
            script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerOptions: ["-v foo:bar", "-u 0"]
        ) {
            bodyExecuted = true
        }

        assertTrue(loggingRule.log.contains('Executing inside a Kubernetes Pod'))
        assertThat(kubernetesConfig.securityContext, is([
            'runAsUser': 0
        ]))
        assertTrue(bodyExecuted)
    }

    @Test
    void testExecuteInsidePodWithCustomUserGroupString() throws Exception {
        Map kubernetesConfig = [:]
        helper.registerAllowedMethod('dockerExecuteOnKubernetes', [Map.class, Closure.class], { Map config, Closure body ->
            kubernetesConfig = config
            return body()
        })
        binding.setVariable('env', [ON_K8S: 'true'])
        stepRule.step.dockerExecute(
            script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerOptions: ["-v foo:bar", "-u root:wheel"]
        ) {
            bodyExecuted = true
        }

        assertTrue(loggingRule.log.contains('Executing inside a Kubernetes Pod'))
        assertThat(kubernetesConfig.securityContext, is([
            'runAsUser': 'root',
            'runAsGroup': 'wheel'
        ]))
        assertTrue(bodyExecuted)
    }

    @Test
    void testExecuteInsideDockerContainer() throws Exception {
        stepRule.step.dockerExecute(script: nullScript, dockerImage: 'maven:3.5-jdk-8-alpine') {
            bodyExecuted = true
        }
        assertEquals('maven:3.5-jdk-8-alpine', docker.getImageNames()[0])
        assertTrue(docker.isImagePulled())
        assertEquals('--env http_proxy --env https_proxy --env no_proxy --env HTTP_PROXY --env HTTPS_PROXY --env NO_PROXY', docker.getParameters().trim())
        assertTrue(bodyExecuted)
    }

    @Test
    void testSkipDockerImagePull() throws Exception {
        nullScript.commonPipelineEnvironment.configuration = [steps: [dockerExecute: [dockerPullImage: false]]]
        stepRule.step.dockerExecute(
            script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine'
        ) {
            bodyExecuted = true
        }
        assertThat(docker.imagePullCount, is(0))
        assertThat(bodyExecuted, is(true))
    }

    @Test
    void testPullSidecarWithDedicatedCredentialsAndRegistry() {
        nullScript.commonPipelineEnvironment.configuration =
        [
            steps: [
                dockerExecute: [
                    dockerRegistryUrl: 'https://registry.example.org',
                    dockerRegistryCredentialsId: 'mySecrets',
                    sidecarRegistryUrl: 'https://sidecarregistry.example.org',
                    sidecarRegistryCredentialsId: 'mySidecarRegistryCredentials',
                ]
            ]
        ]
        stepRule.step.dockerExecute(
            script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerRegistryCredentialsId: 'mySecrets',
            sidecarImage: 'ubuntu',
        ) {
            bodyExecuted = true
        }
        // not clear which image has been pulled with which registry, but at least
        // both registries are involved.
        assertThat(docker.registriesWithCredentials, is([
            [
                registry: 'https://registry.example.org',
                credentialsId: 'mySecrets',
            ],
            [
                registry: 'https://sidecarregistry.example.org',
                credentialsId: 'mySidecarRegistryCredentials',
            ]
        ]))
        assertThat(docker.imagePullCount, is(2))
        assertThat(bodyExecuted, is(true))
    }

    @Test
    void testPullSidecarWithSameCredentialsAndRegistryLikeBaseImageWhenNothingElseIsSpecified() {
        nullScript.commonPipelineEnvironment.configuration =
        [
            steps: [
                dockerExecute: [
                    dockerRegistryUrl: 'https://registry.example.org',
                ]
            ]
        ]
        stepRule.step.dockerExecute(
            script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerRegistryCredentialsId: 'mySecrets',
            sidecarImage: 'ubuntu',
        ) {
            bodyExecuted = true
        }
        // from getting an empty list we derive withRegistry has not been called
        // if it would have been called we would have the registry provided above.
        assertThat(docker.registriesWithCredentials, is([
            [
                registry: 'https://registry.example.org',
                credentialsId: 'mySecrets',
            ],
            [
                registry: 'https://registry.example.org',
                credentialsId: 'mySecrets',
            ],
        ]))
        assertThat(docker.imagePullCount, is(2))
        assertThat(bodyExecuted, is(true))
    }

    @Test
    void testPullWithRegistryOnlyAndNoCredentials() {
        nullScript.commonPipelineEnvironment.configuration =
        [
            steps: [
                dockerExecute: [
                    dockerRegistryUrl: 'https://registry.example.org',
                ]
            ]
        ]
        stepRule.step.dockerExecute(
            script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine'
        ) {
            bodyExecuted = true
        }
        // from getting an empty list we derive withRegistry has not been called
        // if it would have been called we would have the registry provided above.
        assertThat(docker.registriesWithCredentials, is([
            [
                registry: 'https://registry.example.org',
            ]
        ]))
        assertThat(docker.imagePullCount, is(1))
        assertThat(bodyExecuted, is(true))
    }

    @Test
    void testPullWithCredentials() throws Exception {

        nullScript.commonPipelineEnvironment.configuration =
        [
            steps: [
                dockerExecute: [
                    dockerRegistryUrl: 'https://registry.example.org',
                    dockerRegistryCredentialsId: 'mySecrets',
                ]
            ]
        ]
        stepRule.step.dockerExecute(
            script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine'
        ) {
            bodyExecuted = true
        }
        assertThat(docker.registriesWithCredentials, is([
            [
                registry: 'https://registry.example.org',
                credentialsId: 'mySecrets',
            ]
        ]))
        assertThat(docker.imagePullCount, is(1))
        assertThat(bodyExecuted, is(true))
    }

    @Test
    void testSkipSidecarImagePull() throws Exception {
        stepRule.step.dockerExecute(
            script: nullScript,
            dockerName: 'maven',
            dockerImage: 'maven:3.5-jdk-8-alpine',
            sidecarEnvVars: ['testEnv': 'testVal'],
            sidecarImage: 'selenium/standalone-chrome',
            sidecarVolumeBind: ['/dev/shm': '/dev/shm'],
            sidecarName: 'testAlias',
            sidecarPorts: ['4444': '4444', '1111': '1111'],
            sidecarPullImage: false
        ) {
            bodyExecuted = true
        }
        assertThat(docker.imagePullCount, is(1))
        assertThat(bodyExecuted, is(true))
    }

    @Test
    void testExecuteInsideDockerContainerWithParameters() throws Exception {
        stepRule.step.dockerExecute(script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerOptions: '-description=lorem ipsum',
            dockerVolumeBind: ['my_vol': '/my_vol'],
            dockerEnvVars: ['http_proxy': 'http://proxy:8000']) {
            bodyExecuted = true
        }
        assertTrue(docker.getParameters().contains('--env https_proxy '))
        assertTrue(docker.getParameters().contains('--env http_proxy=http://proxy:8000'))
        assertTrue(docker.getParameters().contains('description=lorem\\ ipsum'))
        assertTrue(docker.getParameters().contains('--volume my_vol:/my_vol'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testExecuteInsideDockerContainerWithDockerOptionsList() throws Exception {
        stepRule.step.dockerExecute(script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerOptions: ['-it', '--network=my-network', 'description=lorem ipsum'],
            dockerEnvVars: ['http_proxy': 'http://proxy:8000']) {
            bodyExecuted = true
        }
        assertTrue(docker.getParameters().contains('--env http_proxy=http://proxy:8000'))
        assertTrue(docker.getParameters().contains('-it'))
        assertTrue(docker.getParameters().contains('--network=my-network'))
        assertTrue(docker.getParameters().contains('description=lorem\\ ipsum'))
    }

    @Test
    void testDockerNotInstalledResultsInLocalExecution() throws Exception {
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, "docker .*", 1)
        stepRule.step.dockerExecute(script: nullScript,
            dockerOptions: '-it') {
            bodyExecuted = true
        }
        assertTrue(loggingRule.log.contains('Cannot connect to docker daemon'))
        assertTrue(loggingRule.log.contains('Running on local environment'))
        assertTrue(bodyExecuted)
        assertFalse(docker.isImagePulled())
    }

    @Test
    void testSidecarDefault() {
        stepRule.step.dockerExecute(
            script: nullScript,
            dockerName: 'maven',
            dockerImage: 'maven:3.5-jdk-8-alpine',
            sidecarEnvVars: ['testEnv': 'testVal'],
            sidecarImage: 'selenium/standalone-chrome',
            sidecarVolumeBind: ['/dev/shm': '/dev/shm'],
            sidecarName: 'testAlias',
            sidecarPorts: ['4444': '4444', '1111': '1111']
        ) {
            bodyExecuted = true
        }

        assertThat(bodyExecuted, is(true))
        assertThat(docker.imagePullCount, is(2))
        assertThat(docker.sidecarParameters, allOf(
            containsString('--env testEnv=testVal'),
            containsString('--volume /dev/shm:/dev/shm'),
            containsString('--network sidecar-'),
            containsString('--network-alias testAlias')
        ))
        assertThat(docker.parameters, allOf(
            containsString('--network sidecar-'),
            containsString('--network-alias maven')
        ))
    }

    @Test
    void testSidecarHealthCheck() {
        stepRule.step.dockerExecute(
            script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            sidecarImage: 'selenium/standalone-chrome',
            sidecarName: 'testAlias',
            sidecarReadyCommand: "isReady.sh"
        ) {}
        assertThat(shellRule.shell, hasItem("docker exec uniqueId isReady.sh"))
    }

    @Test
    void testSidecarKubernetes() {
        boolean dockerExecuteOnKubernetesCalled = false
        binding.setVariable('env', [ON_K8S: 'true'])
        helper.registerAllowedMethod('dockerExecuteOnKubernetes', [Map.class, Closure.class], { params, body ->
            dockerExecuteOnKubernetesCalled = true
            assertThat(params.dockerImage, is('maven:3.5-jdk-8-alpine'))
            assertThat(params.containerName, is('maven'))
            assertThat(params.sidecarEnvVars, is(['testEnv': 'testVal']))
            assertThat(params.sidecarName, is('selenium'))
            assertThat(params.sidecarImage, is('selenium/standalone-chrome'))
            assertThat(params.containerName, is('maven'))
            assertThat(params.containerPortMappings['selenium/standalone-chrome'], hasItem(allOf(hasEntry('containerPort', 4444), hasEntry('hostPort', 4444))))
            assertThat(params.dockerWorkspace, is('/home/piper'))
            body()
        })
        stepRule.step.dockerExecute(
            script: nullScript,
            containerPortMappings: [
                'selenium/standalone-chrome': [[name: 'selPort', containerPort: 4444, hostPort: 4444]]
            ],
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerName: 'maven',
            dockerWorkspace: '/home/piper',
            sidecarEnvVars: ['testEnv': 'testVal'],
            sidecarImage: 'selenium/standalone-chrome',
            sidecarName: 'selenium',
            sidecarVolumeBind: ['/dev/shm': '/dev/shm']
        ) {
            bodyExecuted = true
        }
        assertThat(bodyExecuted, is(true))
        assertThat(dockerExecuteOnKubernetesCalled, is(true))
    }

    @Test
    void testSidecarKubernetesHealthCheck() {
        binding.setVariable('env', [ON_K8S: 'true'])

        helper.registerAllowedMethod('dockerExecuteOnKubernetes', [Map.class, Closure.class], { params, body ->
            body()
            SidecarUtils sidecarUtils = new SidecarUtils(nullScript)
            sidecarUtils.waitForSidecarReadyOnKubernetes(params.sidecarName, params.sidecarReadyCommand)
        })

        def containerCalled = false
        helper.registerAllowedMethod('container', [Map.class, Closure.class], { params, body ->
            containerCalled = true
            assertThat(params.name, is('testAlias'))
            body()
        })

        stepRule.step.dockerExecute(
            script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            sidecarImage: 'selenium/standalone-chrome',
            sidecarName: 'testAlias',
            sidecarReadyCommand: "isReady.sh"
        ) {}

        assertThat(containerCalled, is(true))
        assertThat(shellRule.shell, hasItem("isReady.sh"))
    }

    private class DockerMock {
        private List imageNames = []
        private boolean imagePulled = false
        private int imagePullCount = 0
        private String parameters
        private String sidecarParameters
        private List registriesWithCredentials = []
        private String credentialsId

        DockerMock image(String imageName) {
            this.imageNames << imageName
            return this
        }

        void pull() {
            imagePullCount++
            imagePulled = true
        }

        void withRegistry(String  registry, String credentialsId, Closure c) {
            this.registriesWithCredentials << [registry: registry, credentialsId: credentialsId]
            c()
        }

        void withRegistry(String  registry, Closure c) {
            this.registriesWithCredentials << [registry: registry]
            c()
        }

        void inside(String parameters, body) {
            this.parameters = parameters
            body()
        }

        void withRun(String parameters, body) {
            this.sidecarParameters = parameters
            body([id: 'uniqueId'])
        }

        def getImageNames() {
            return imageNames
        }

        boolean isImagePulled() {
            return imagePulled
        }

        String getParameters() {
            return parameters
        }
    }
}
