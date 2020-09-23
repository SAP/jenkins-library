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

import hudson.AbortException

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertFalse
import static org.junit.Assert.assertNull

class DockerExecuteOnKubernetesTest extends BasePiperTest {
    private ExpectedException exception = ExpectedException.none()
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private ExpectedException thrown = new ExpectedException()

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(exception)
        .around(shellRule)
        .around(loggingRule)
        .around(stepRule)
        .around(thrown)
    int whichDockerReturnValue = 0
    def bodyExecuted
    def dockerImage
    def containerMap
    def dockerEnvVars
    def dockerWorkspace
    def podName = ''
    def podLabel = ''
    def podNodeSelector = ''
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
    def inheritFrom
    def yamlMergeStrategy
    Map resources =  [:]
    List stashList = []

    @Before
    void init() {
        containersList = []
        imageList = []
        envList = []
        portList = []
        containerCommands = []
        resources = [:]
        bodyExecuted = false
        JenkinsUtils.metaClass.static.isPluginActive = { def s -> new PluginMock(s).isActive() }
        helper.registerAllowedMethod('sh', [Map.class], { return whichDockerReturnValue })
        helper.registerAllowedMethod('container', [Map.class, Closure.class], { Map config, Closure body ->
            container(config) {
                body()
            }
        })
        helper.registerAllowedMethod('merge', [], {
            return 'merge'
        })
        helper.registerAllowedMethod('podTemplate', [Map.class, Closure.class], { Map options, Closure body ->
            podName = options.name
            podLabel = options.label
            namespace = options.namespace
            inheritFrom = options.inheritFrom
            yamlMergeStrategy = options.yamlMergeStrategy
            podNodeSelector = options.nodeSelector
            def podSpec = new JsonSlurper().parseText(options.yaml)  // this yaml is actually json
            def containers = podSpec.spec.containers
            securityContext = podSpec.spec.securityContext

            containers.each { container ->
                containersList.add(container.name)
                imageList.add(container.image.toString())
                envList.add(container.env)
                if (container.ports) {
                    portList.add(container.ports)
                }
                if (container.command) {
                    containerCommands.add(container.command)
                }
                pullImageMap.put(container.image.toString(), container.imagePullPolicy == "Always")
                resources.put(container.name, container.resources)
            }
            body()
        })
        helper.registerAllowedMethod('stash', [Map.class], { m ->
            stashList.add(m)
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
        ) {
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
    void testDockerExecuteOnKubernetesEmptyContainerMapNoDockerImage() throws Exception {
        stepRule.step.dockerExecuteOnKubernetes(
            script: nullScript,
            juStabUtils: utils,
            containerMap: [:],
            dockerEnvVars: ['customEnvKey': 'customEnvValue']) {
            bodyExecuted = true
        }
        assertTrue(bodyExecuted)
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
    void testInheritFromPodTemplate() throws Exception {
        nullScript.commonPipelineEnvironment.configuration = ['general': ['jenkinsKubernetes': ['inheritFrom': 'default']]]
        stepRule.step.dockerExecuteOnKubernetes(script: nullScript,
            containerMap: ['maven:3.5-jdk-8-alpine': 'mavenexecute']) {
            container(name: 'mavenexecute') {
                bodyExecuted = true
            }
        }
        assertEquals(inheritFrom, 'default')
        assertEquals(yamlMergeStrategy, 'merge')
        assertTrue(bodyExecuted)
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
    void testDockerExecuteOnKubernetesNoResourceLimitsOnEmptyResourcesMap() throws Exception {

        nullScript.commonPipelineEnvironment.configuration = [general:
            [jenkinsKubernetes: [
                resources: [
                    DEFAULT: [
                        requests: [
                            memory: '1Gi',
                            cpu: '0.25'
                        ],
                        limits: [
                            memory: '2Gi',
                            cpu: '1'
                        ]
                    ],
                    mavenexecute: [:]
                ]
            ]
        ]]
        stepRule.step.dockerExecuteOnKubernetes(script: nullScript,
            containerMap: ['maven:3.5-jdk-8-alpine': 'mavenexecute'], {
                bodyExecuted = true
            })

        assertNull(resources.mavenexecute)
        assertTrue(bodyExecuted)
    }

    @Test
    void testDockerExecuteOnKubernetesWithDefaultResourceLimits() throws Exception {

        nullScript.commonPipelineEnvironment.configuration = [general:
            [jenkinsKubernetes: [
                resources: [DEFAULT: [
                    requests: [
                        memory: '1Gi',
                        cpu: '0.25'
                    ],
                    limits: [
                        memory: '2Gi',
                        cpu: '1'
                    ]
                ]
            ]
        ]]]
        stepRule.step.dockerExecuteOnKubernetes(script: nullScript,
            containerMap: ['maven:3.5-jdk-8-alpine': 'mavenexecute'], {
                bodyExecuted = true
            })

        assertEquals(requests: [memory: '1Gi',cpu: '0.25'],limits: [memory: '2Gi',cpu: '1'], resources.jnlp)
        assertEquals(requests: [memory: '1Gi',cpu: '0.25'],limits: [memory: '2Gi',cpu: '1'], resources.mavenexecute)
        assertTrue(bodyExecuted)
    }

    @Test
    void testDockerExecuteOnKubernetesWithSpecificResourcLimitsParametersAreTakingPrecendence() throws Exception {

        // the settings here are expected to be overwritten by the parameters provided via signature
        nullScript.commonPipelineEnvironment.configuration = [general:
            [jenkinsKubernetes: [
                resources: [
                    mavenexecute: [
                    requests: [
                        memory: '2Gi',
                        cpu: '0.75'
                    ],
                    limits: [
                        memory: '4Gi',
                        cpu: '2'
                    ]
                ]
            ]
        ]]]
        stepRule.step.dockerExecuteOnKubernetes(script: nullScript,
            containerMap: ['maven:3.5-jdk-8-alpine': 'mavenexecute'],
            resources: [
                    mavenexecute: [
                    requests: [
                        memory: '8Gi',
                        cpu: '2'
                    ],
                    limits: [
                        memory: '16Gi',
                        cpu: '4'
                    ]
                ]
            ]) {
                bodyExecuted = true
            }

        assertEquals(requests: [memory: '8Gi',cpu: '2'],limits: [memory: '16Gi',cpu: '4'], resources.mavenexecute)
        assertTrue(bodyExecuted)
    }

    @Test
    void testDockerExecuteOnKubernetesWithSpecificResourceLimits() throws Exception {

        nullScript.commonPipelineEnvironment.configuration = [general:
            [jenkinsKubernetes: [
                resources: [
                    DEFAULT: [
                        requests: [
                            memory: '1Gi',
                            cpu: '0.25'
                        ],
                        limits: [
                            memory: '2Gi',
                            cpu: '1'
                        ]
                    ],
                    mavenexecute: [
                        requests: [
                            memory: '2Gi',
                            cpu: '0.75'
                        ],
                        limits: [
                            memory: '4Gi',
                            cpu: '2'
                        ]
                    ],
                    jnlp: [
                        requests: [
                            memory: '3Gi',
                            cpu: '0.33'
                        ],
                        limits: [
                            memory: '6Gi',
                            cpu: '3'
                        ]
                    ],
                    mysidecar: [
                        requests: [
                            memory: '10Gi',
                            cpu: '5.00'
                        ],
                        limits: [
                            memory: '20Gi',
                            cpu: '10'
                        ]
                    ]
                ]
            ]
        ]]
        stepRule.step.dockerExecuteOnKubernetes(script: nullScript,
            containerMap: ['maven:3.5-jdk-8-alpine': 'mavenexecute'],
            sidecarImage: 'ubuntu',
            sidecarName: 'mysidecar') {
                bodyExecuted = true
            }
            
        assertEquals(requests: [memory: '10Gi',cpu: '5.00'],limits: [memory: '20Gi',cpu: '10'], resources.mysidecar)
        assertEquals(requests: [memory: '3Gi',cpu: '0.33'],limits: [memory: '6Gi',cpu: '3'], resources.jnlp)
        assertEquals(requests: [memory: '2Gi',cpu: '0.75'],limits: [memory: '4Gi',cpu: '2'], resources.mavenexecute)
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
    void testSidecarDefaultWithContainerMap() {
        List portMapping = []
        helper.registerAllowedMethod('portMapping', [Map.class], { m ->
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
                'maven:3.5-jdk-8-alpine'    : 'mavenexecute',
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
            hasItem('mavenexecute'),
            hasItem('selenium'),
        ))
        assertThat(imageList, allOf(
            hasItem('maven:3.5-jdk-8-alpine'),
            hasItem('selenium/standalone-chrome'),
        ))
        assertThat(portList, hasItem([[name: 'selenium0', containerPort: 4444]]))
        assertThat(containerCommands.size(), is(1))
        assertThat(envList, hasItem(hasItem(allOf(hasEntry('name', 'customEnvKey'), hasEntry('value', 'customEnvValue')))))
    }

    @Test
    void testSidecarDefaultWithParameters() {
        List portMapping = []
        helper.registerAllowedMethod('portMapping', [Map.class], { m ->
            portMapping.add(m)
            return m
        })
        stepRule.step.dockerExecuteOnKubernetes(
            script: nullScript,
            juStabUtils: utils,
            containerMap: ['maven:3.5-jdk-8-alpine': 'mavenexecute'],
            containerName: 'mavenexecute',
            dockerOptions: '-it',
            dockerVolumeBind: ['my_vol': '/my_vol'],
            dockerEnvVars: ['http_proxy': 'http://proxy:8000'],
            dockerWorkspace: '/home/piper',
            sidecarEnvVars: ['testEnv': 'testVal'],
            sidecarWorkspace: '/home/piper/sidecar',
            sidecarImage: 'postgres',
            sidecarName: 'postgres',
            sidecarReadyCommand: 'pg_isready'
        ) {
            bodyExecuted = true
        }

        assertThat(bodyExecuted, is(true))

        assertThat(containersList, allOf(hasItem('postgres'), hasItem('mavenexecute')))
        assertThat(imageList, allOf(hasItem('maven:3.5-jdk-8-alpine'), hasItem('postgres')))

        assertThat(envList, hasItem(hasItem(allOf(hasEntry('name', 'testEnv'), hasEntry('value', 'testVal')))))
        assertThat(envList, hasItem(hasItem(allOf(hasEntry('name', 'HOME'), hasEntry('value', '/home/piper/sidecar')))))
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
                'maven:3.5-jdk-8-alpine'    : 'mavenexecute',
                'selenium/standalone-chrome': 'selenium'
            ],
            containerName: 'mavenexecute',
            containerWorkspaces: [
                'selenium/standalone-chrome': ''
            ],
            containerPullImageFlags: [
                'maven:3.5-jdk-8-alpine'    : true,
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
        def expectedSecurityContext = [runAsUser: 1000, fsGroup: 1000]
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

    @Test
    void testDockerExecuteOnKubernetesCustomNode() {

        stepRule.step.dockerExecuteOnKubernetes(
            script: nullScript,
            juStabUtils: utils,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            nodeSelector: 'size:big'
        ) { bodyExecuted = true }
        assertTrue(bodyExecuted)
        assertThat(podNodeSelector, is('size:big'))
    }

    @Test
    void testDockerExecuteOnKubernetesCustomJnlpViaEnv() {

        nullScript.commonPipelineEnvironment.configuration = [
            general: [jenkinsKubernetes: [jnlpAgent: 'config/jnlp:latest']]
        ]
        binding.variables.env.JENKINS_JNLP_IMAGE = 'env/jnlp:latest'
        stepRule.step.dockerExecuteOnKubernetes(
            script: nullScript,
            juStabUtils: utils,
            dockerImage: 'maven:3.5-jdk-8-alpine',
        ) { bodyExecuted = true }
        assertTrue(bodyExecuted)

        assertThat(containersList, allOf(
            hasItem('jnlp'),
            hasItem('container-exec')
        ))
        assertThat(imageList, allOf(
            hasItem('env/jnlp:latest'),
            hasItem('maven:3.5-jdk-8-alpine'),
        ))
    }

    @Test
    void testDockerExecuteOnKubernetesCustomJnlpViaConfig() {

        nullScript.commonPipelineEnvironment.configuration = [
            general: [jenkinsKubernetes: [jnlpAgent: 'config/jnlp:latest']]
        ]
        //binding.variables.env.JENKINS_JNLP_IMAGE = 'config/jnlp:latest'
        stepRule.step.dockerExecuteOnKubernetes(
            script: nullScript,
            juStabUtils: utils,
            dockerImage: 'maven:3.5-jdk-8-alpine',
        ) { bodyExecuted = true }
        assertTrue(bodyExecuted)

        assertThat(containersList, allOf(
            hasItem('jnlp'),
            hasItem('container-exec')
        ))
        assertThat(imageList, allOf(
            hasItem('config/jnlp:latest'),
            hasItem('maven:3.5-jdk-8-alpine'),
        ))
    }

    @Test
    void testDockerExecuteOnKubernetesExecutionFails() {

        thrown.expect(AbortException)
        thrown.expectMessage('Execution failed.')

        nullScript.commonPipelineEnvironment.configuration = [
            general: [jenkinsKubernetes: [jnlpAgent: 'config/jnlp:latest']]
        ]
        //binding.variables.env.JENKINS_JNLP_IMAGE = 'config/jnlp:latest'
        stepRule.step.dockerExecuteOnKubernetes(
            script: nullScript,
            juStabUtils: utils,
            dockerImage: 'maven:3.5-jdk-8-alpine',
        ) { throw new AbortException('Execution failed.') }
    }

    @Test
    void testStashIncludesAndExcludes() {
        nullScript.commonPipelineEnvironment.configuration = [
            steps: [
                dockerExecuteOnKubernetes: [
                    stashExcludes: [
                        workspace: 'workspace/exclude.test',
                        stashBack: 'container/exclude.test'
                    ],
                    stashIncludes: [
                        workspace: 'workspace/include.test',
                        stashBack: 'container/include.test'
                    ]
                ]
            ]
        ]
        stepRule.step.dockerExecuteOnKubernetes(
            script: nullScript,
            juStabUtils: utils,
            dockerImage: 'maven:3.5-jdk-8-alpine',
        ) {
            bodyExecuted = true
        }
        assertThat(stashList, hasItem(allOf(
            not(hasEntry('allowEmpty', true)),
            hasEntry('includes', 'workspace/include.test'),
            hasEntry('excludes', 'workspace/exclude.test'))))
        assertThat(stashList, hasItem(allOf(
            not(hasEntry('allowEmpty', true)),
            hasEntry('includes', 'container/include.test'),
            hasEntry('excludes', 'container/exclude.test'))))
    }


    private container(options, body) {
        containerName = options.name
        containerShell = options.shell
        body()
    }
}
