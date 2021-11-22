package com.sap.piper.k8s

import com.sap.piper.DefaultValueCache
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.hasItem
import static org.junit.Assert.*

class ContainerMapTest extends BasePiperTest {

    String exampleConfigYaml = """
# Mapping of Go step names to their YAML metadata resource file
stepMetadata:
  artifactPrepareVersion: artifactPrepareVersion.yaml
containerMaps:
  init:
    - artifactPrepareVersion
    - mavenExecute
"""

    private JenkinsShellCallRule shellCallRule = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(new JenkinsPiperExecuteBinRule(this))
        .around(new JenkinsWriteFileRule(this))
        .around(new JenkinsReadJsonRule(this))
        .around(shellCallRule)

    private List envs

    @Before
    void init() {
        helper.registerAllowedMethod('libraryResource', [String.class], {
            return exampleConfigYaml
        })

        helper.registerAllowedMethod('withEnv', [List.class, Closure.class], { List envs, Closure body ->
            this.envs = envs.collect {it.toString()}
            body()
        })

        DefaultValueCache.createInstance([
            steps: [
                mavenExecute: [
                    dockerImage: 'maven:3.5-jdk-8-alpine'
                ]
            ]
        ])

        shellCallRule.setReturnValue('./piper getConfig --contextConfig --stepMetadata \'.pipeline/tmp/metadata/artifactPrepareVersion.yaml\'', '{"dockerImage":"artifact-image"}')
    }

    @Test
    void testIfObjectCreated() {
        assertNotNull(ContainerMap.instance)
    }

    @Test
    void testSetMap() {
        ContainerMap.instance.setMap(['testpod': ['maven:3.5-jdk-8-alpine': 'mavenexec']])
        assertEquals(['testpod': ['maven:3.5-jdk-8-alpine': 'mavenexec']],ContainerMap.instance.getMap())
    }

    @Test
    void testGetMap() {
        assertNotNull(ContainerMap.instance.getMap())
    }

    @Test
    void testInitFromResource(){
        ContainerMap.instance.initFromResource(nullScript, 'containersMapYaml', 'maven', utils)
        assertThat(envs, hasItem('STAGE_NAME=init'))
        assertThat(envs, hasItem('PIPER_parametersJSON={"buildTool":"maven"}'))

        assertEquals(['init': ['artifact-image':'artifactprepareversion', 'maven:3.5-jdk-8-alpine': 'mavenexecute']],ContainerMap.instance.getMap())
    }
}
