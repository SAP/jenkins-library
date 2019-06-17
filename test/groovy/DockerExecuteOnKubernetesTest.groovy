import com.sap.piper.JenkinsUtils

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain


import groovy.json.JsonSlurper
import util.BasePiperTest
import util.JenkinsDockerExecuteRule
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

class DockerExecuteOnKubernetesTest extends BasePiperTest {
    private ExpectedException exception = ExpectedException.none()
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(exception)
        .around(dockerExecuteRule)
        .around(shellRule)
        .around(loggingRule)
        .around(stepRule)
    int whichDockerReturnValue = 0
    def bodyExecuted
    def dockerImage
    def containerMap
    def dockerEnvVars
    def dockerWorkspace
    def podName = ''
    def podLabel = ''
    def containersList = []
    def imageList = []
    def containerName = ''
    def containerShell = ''
    def envList = []
    def portList = []
    def containerCommands = []
    def pullImageMap = [:]
    def namespace
    def securityContext
    Map stashMap

    @Before
    void init() {
        containersList = []
        imageList = []
        envList = []
        portList = []
        containerCommands = []
        bodyExecuted = false
        JenkinsUtils.metaClass.static.isPluginActive = { def s -> new PluginMock(s).isActive() }
        helper.registerAllowedMethod('sh', [Map.class], {return whichDockerReturnValue})
        helper.registerAllowedMethod('container', [Map.class, Closure.class], { Map config, Closure body -> container(config){body()}
        })
        helper.registerAllowedMethod('podTemplate', [Map.class, Closure.class], { Map options, Closure body ->
            podName = options.name
            podLabel = options.label
            namespace = options.namespace
            def podSpec = new JsonSlurper().parseText(options.yaml)  // this yaml is actually json
            def containers = podSpec.spec.containers
            securityContext = podSpec.spec.securityContext

            containers.each { container ->
                containersList.add(container.name)
                imageList.add(container.image.toString())
                envList.add(container.env)
                if(container.ports) {
                    portList.add(container.ports)
                }
                if (container.command) {
                    containerCommands.add(container.command)
                }
                pullImageMap.put(container.image.toString(), container.imagePullPolicy == "Always")
            }
            body()
        })
        helper.registerAllowedMethod('stash', [Map.class], {m ->
            stashMap = m
        })

    }

    @Test
    void testRunOnPodNoContainerMapOnlyDockerImage() throws Exception {
        stepRule.step.dockerExecuteOnKubernetes(
            script: nullScript,
            juStabUtils: utils,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            dockerOptions: '-it',
            dockerVolumeBind: ['my_vol': '/my_vol'],
            dockerEnvVars: ['http_proxy': 'http://proxy:8000'], dockerWorkspace: '/home/piper'
        ){
            bodyExecuted = true
        }
        assertThat(containersList, hasItem('container-exec'))
        assertThat(imageList, hasItem('maven:3.5-jdk-8-alpine'))
        assertThat(envList.toString(), containsString('http_proxy'))
        assertThat(envList.toString(), containsString('http://proxy:8000'))
        assertThat(envList.toString(), containsString('/home/piper'))
        assertThat(bodyExecuted, is(true))
        assertThat(containerCommands.size(), is(1))
    }

    @Test
    void testDockerExecuteOnKubernetesWithCustomContainerMap() throws Exception {
        stepRule.step.dockerExecuteOnKubernetes(script: nullScript,
            containerMap: ['maven:3.5-jdk-8-alpine': 'mavenexecute']) {
            container(name: 'mavenexecute') {
                bodyExecuted = true
            }
        }
        assertEquals('mavenexecute', containerName)
        assertTrue(containersList.contains('mavenexecute'))
        assertTrue(imageList.contains('maven:3.5-jdk-8-alpine'))
        assertTrue(bodyExecuted)
        assertThat(containerCommands.size(), is(1))
    }

    @Test
    void testDockerExecuteOnKubernetesWithCustomJnlpWithContainerMap() throws Exception {
        nullScript.commonPipelineEnvironment.configuration = ['general': ['jenkinsKubernetes': ['jnlpAgent': 'myJnalpAgent']]]
        stepRule.step.dockerExecuteOnKubernetes(script: nullScript,
            containerMap: ['maven:3.5-jdk-8-alpine': 'mavenexecute']) {
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
    void testDockerExecuteOnKubernetesWithCustomJnlpWithDockerImage() throws Exception {
        nullScript.commonPipelineEnvironment.configuration = ['general': ['jenkinsKubernetes': ['jnlpAgent': 'myJnalpAgent']]]
        stepRule.step.dockerExecuteOnKubernetes(
            script: nullScript,
            juStabUtils: utils,
            dockerImage: 'maven:3.5-jdk-8-alpine') {
            bodyExecuted = true
        }
        assertEquals('container-exec', containerName)
        assertTrue(containersList.contains('jnlp'))
        assertTrue(containersList.contains('container-exec'))
        assertTrue(imageList.contains('myJnalpAgent'))
        assertTrue(imageList.contains('maven:3.5-jdk-8-alpine'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testDockerExecuteOnKubernetesWithCustomWorkspace() throws Exception {
        stepRule.step.dockerExecuteOnKubernetes(script: nullScript,
            containerMap: ['maven:3.5-jdk-8-alpine': 'mavenexecute'],
            dockerWorkspace: '/home/piper') {
            container(name: 'mavenexecute') {
                bodyExecuted = true
            }
        }
        assertTrue(envList.toString().contains('/home/piper'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testDockerExecuteOnKubernetesWithCustomEnv() throws Exception {
        stepRule.step.dockerExecuteOnKubernetes(script: nullScript,
            containerMap: ['maven:3.5-jdk-8-alpine': 'mavenexecute'],
            dockerEnvVars: ['customEnvKey': 'customEnvValue']) {
            container(name: 'mavenexecute') {
                bodyExecuted = true
            }
        }
        assertTrue(envList.toString().contains('customEnvKey') && envList.toString().contains('customEnvValue'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testDockerExecuteOnKubernetesUpperCaseContainerName() throws Exception {
        stepRule.step.dockerExecuteOnKubernetes(script: nullScript,
            containerMap: ['maven:3.5-jdk-8-alpine': 'MAVENEXECUTE'],
            dockerEnvVars: ['customEnvKey': 'customEnvValue']) {
            container(name: 'mavenexecute') {
                bodyExecuted = true
            }
        }
        assertEquals('mavenexecute', containerName)
        assertTrue(containersList.contains('mavenexecute'))
        assertTrue(imageList.contains('maven:3.5-jdk-8-alpine'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testDockerExecuteOnKubernetesEmptyContainerMapNoDockerImage() throws Exception {
        exception.expect(IllegalArgumentException.class)
            stepRule.step.dockerExecuteOnKubernetes(
                script: nullScript,
                juStabUtils: utils,
                containerMap: [:],
                dockerEnvVars: ['customEnvKey': 'customEnvValue']) {
                container(name: 'jnlp') {
                    bodyExecuted = true
                }
            }
        assertFalse(bodyExecuted)
    }

    @Test
    void testSidecarDefault() {
        List portMapping = []
        helper.registerAllowedMethod('portMapping', [Map.class], {m ->
            portMapping.add(m)
            return m
        })
        stepRule.step.dockerExecuteOnKubernetes(
            script: nullScript,
            juStabUtils: utils,
            containerCommands: ['selenium/standalone-chrome': ''],
            containerEnvVars: [
                'selenium/standalone-chrome': ['customEnvKey': 'customEnvValue']
            ],
            containerMap: [
                'maven:3.5-jdk-8-alpine': 'mavenexecute',
                'selenium/standalone-chrome': 'selenium'
            ],
            containerName: 'mavenexecute',
            containerPortMappings: [
                'selenium/standalone-chrome': [[containerPort: 4444]]
            ],
            containerWorkspaces: [
                'selenium/standalone-chrome': ''
            ],
            dockerWorkspace: '/home/piper'
        ) {
            bodyExecuted = true
        }

        assertThat(bodyExecuted, is(true))
        assertThat(containerName, is('mavenexecute'))

        assertThat(containersList, allOf(
            hasItem('jnlp'),
            hasItem('mavenexecute'),
            hasItem('selenium'),
        ))
        assertThat(imageList, allOf(
            hasItem('s4sdk/jenkins-agent-k8s:latest'),
            hasItem('maven:3.5-jdk-8-alpine'),
            hasItem('selenium/standalone-chrome'),
        ))
        assertThat(portList, hasItem([[name: 'selenium0', containerPort: 4444]]))
        assertThat(containerCommands.size(), is(1))
        assertThat(envList, hasItem(hasItem(allOf(hasEntry('name', 'customEnvKey'), hasEntry ('value','customEnvValue')))))
    }

    @Test
    void testDockerExecuteOnKubernetesWithCustomShell() {
        stepRule.step.dockerExecuteOnKubernetes(
            script: nullScript,
            juStabUtils: utils,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            containerShell: '/busybox/sh'
        ) {
            //nothing to exeute
        }
        assertThat(containerShell, is('/busybox/sh'))
    }

    @Test
    void testDockerExecuteOnKubernetesWithCustomContainerCommand() {
        stepRule.step.dockerExecuteOnKubernetes(
            script: nullScript,
            juStabUtils: utils,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            containerCommand: '/busybox/tail -f /dev/null'
        ) {
            //nothing to exeute
        }
        assertThat(containerCommands, hasItem(['/bin/sh', '-c', '/busybox/tail -f /dev/null']))
    }

    @Test
    void testSkipDockerImagePull() throws Exception {
        stepRule.step.dockerExecuteOnKubernetes(
            script: nullScript,
            dockerPullImage: false,
            containerMap: ['maven:3.5-jdk-8-alpine': 'mavenexecute']
        ) {
            container(name: 'mavenexecute') {
                bodyExecuted = true
            }
        }
        assertEquals(false, pullImageMap.get('maven:3.5-jdk-8-alpine'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testSkipSidecarImagePull() throws Exception {
        stepRule.step.dockerExecuteOnKubernetes(
            script: nullScript,
            juStabUtils: utils,
            containerCommands: ['selenium/standalone-chrome': ''],
            containerEnvVars: [
                'selenium/standalone-chrome': ['customEnvKey': 'customEnvValue']
            ],
            containerMap: [
                'maven:3.5-jdk-8-alpine': 'mavenexecute',
                'selenium/standalone-chrome': 'selenium'
            ],
            containerName: 'mavenexecute',
            containerWorkspaces: [
                'selenium/standalone-chrome': ''
            ],
            containerPullImageFlags: [
                'maven:3.5-jdk-8-alpine': true,
                'selenium/standalone-chrome': false
            ],
            dockerWorkspace: '/home/piper'
        ) {
            bodyExecuted = true
        }
        assertEquals(true, pullImageMap.get('maven:3.5-jdk-8-alpine'))
        assertEquals(false, pullImageMap.get('selenium/standalone-chrome'))
        assertTrue(bodyExecuted)
    }

    @Test
    void testDockerExecuteOnKubernetesWithCustomNamespace() {
        def expectedNamespace = "sandbox"
        nullScript.commonPipelineEnvironment.configuration = [general: [jenkinsKubernetes: [namespace: expectedNamespace]]]

        stepRule.step.dockerExecuteOnKubernetes(
                script: nullScript,
                juStabUtils: utils,
                dockerImage: 'maven:3.5-jdk-8-alpine',
                ) { bodyExecuted = true }
        assertTrue(bodyExecuted)
        assertThat(namespace, is(equalTo(expectedNamespace)))
    }

    @Test
    void testDockerExecuteOnKubernetesWithSecurityContext() {
        def expectedSecurityContext = [ runAsUser: 1000, fsGroup: 1000 ]
        nullScript.commonPipelineEnvironment.configuration = [general: [jenkinsKubernetes: [
                    securityContext: expectedSecurityContext]]]

        stepRule.step.dockerExecuteOnKubernetes(
                script: nullScript,
                juStabUtils: utils,
                dockerImage: 'maven:3.5-jdk-8-alpine',
                ) { bodyExecuted = true }
        assertTrue(bodyExecuted)
        assertThat(securityContext, is(equalTo(expectedSecurityContext)))
    }

    /*
    Due to negative side-effect of full git stashing
    @Test
    void testDockerExecuteOnKubernetesWorkspaceStashing() {

        stepRule.step.dockerExecuteOnKubernetes(
            script: nullScript,
            juStabUtils: utils,
            dockerImage: 'maven:3.5-jdk-8-alpine',
        ) { bodyExecuted = true }
        assertTrue(bodyExecuted)
        assertThat(stashMap.useDefaultExcludes, is(false))
    }
    */


    private container(options, body) {
        containerName = options.name
        containerShell = options.shell
        body()
    }
}
