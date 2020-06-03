#!groovy
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

import static org.hamcrest.CoreMatchers.containsString
import static org.hamcrest.CoreMatchers.hasItem
import static org.hamcrest.CoreMatchers.is
import static org.hamcrest.CoreMatchers.nullValue
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertFalse
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue

class BuildExecuteTest extends BasePiperTest {
    private ExpectedException exception = ExpectedException.none()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsDockerExecuteRule dockerRule = new JenkinsDockerExecuteRule(this)
    private JenkinsShellCallRule shellCallRule = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(exception)
        .around(shellCallRule)
        .around(dockerRule)
        .around(stepRule)

    def dockerMockArgs = [:]
    class DockerMock {
        DockerMock(name) {
            dockerMockArgs.name = name
        }

        def build(image, options) {
            return [image: image, options: options]
        }
    }

    @Before
    void init() {
    }

    @Test
    void testDefaultError() {
        exception.expectMessage(containsString('buildTool not set and no dockerImage & dockerCommand provided'))
        stepRule.step.buildExecute(
            script: nullScript,
        )
    }

    @Test
    void testDefaultWithDockerImage() {
        stepRule.step.buildExecute(
            script: nullScript,
            dockerImage: 'path/to/myImage:tag',
            dockerCommand: 'myTestCommand'
        )
        assertThat(dockerRule.dockerParams.dockerImage, is('path/to/myImage:tag'))
        assertThat(shellCallRule.shell, hasItem('myTestCommand'))
    }

    @Test
    void testMaven() {
        def buildToolCalled = false
        helper.registerAllowedMethod('mavenBuild', [Map.class], { m ->
            buildToolCalled = true
            return
        })
        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'maven',
        )
        assertThat(buildToolCalled, is(true))
    }

    @Test
    void testMta() {
        def buildToolCalled = false
        helper.registerAllowedMethod('mtaBuild', [Map.class], { m ->
            buildToolCalled = true
            return
        })
        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'mta',
        )
        assertThat(buildToolCalled, is(true))
    }

    @Test
    void testNpm() {
        def buildToolCalled = false
        helper.registerAllowedMethod('npmExecuteScripts', [Map.class], { m ->
            buildToolCalled = true
            return
        })
        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'npm',
        )
        assertThat(buildToolCalled, is(true))
    }

    @Test
    void testNpmWithScripts() {
        boolean actualValue = false
        helper.registerAllowedMethod('npmExecuteScripts', [Map.class], { m ->
            actualValue = (m['runScripts'][0] == 'foo' && m['runScripts'][1] == 'bar')
            return
        })
        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'npm',
            npmRunScripts: ['foo', 'bar']
        )
        assertTrue(actualValue)
    }

    @Test
    void testNpmWithInstallFalse() {
        boolean actualValue = true
        helper.registerAllowedMethod('npmExecuteScripts', [Map.class], { m ->
            actualValue = m['install']
            return
        })
        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'npm',
            npmInstall: false
        )
        assertFalse(actualValue)
    }

    @Test
    void testNpmWithInstallTrue() {
        boolean actualValue = false
        helper.registerAllowedMethod('npmExecuteScripts', [Map.class], { m ->
            actualValue = m['install']
            return
        })
        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'npm',
            npmInstall: true
        )
        assertTrue(actualValue)
    }

    @Test
    void testDocker() {
        binding.setVariable('docker', new DockerMock('test'))
        def pushParams = [:]
        helper.registerAllowedMethod('containerPushToRegistry', [Map.class], { m ->
            pushParams = m
            return
        })
        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'docker',
            dockerImageName: 'path/to/myImage',
            dockerImageTag: 'myTag',
            dockerRegistryUrl: 'https://my.registry:55555'
        )

        assertThat(pushParams.dockerBuildImage.image.toString(), is('path/to/myImage:myTag'))
        assertThat(pushParams.dockerRegistryUrl.toString(), is('https://my.registry:55555'))
        assertThat(nullScript.commonPipelineEnvironment.getValue('containerImage').toString(), is('path/to/myImage:myTag'))
    }

    @Test
    void testDockerWithEnv() {
        nullScript.commonPipelineEnvironment.setArtifactVersion('1.0.0')
        binding.setVariable('docker', new DockerMock('test'))
        def pushParams = [:]
        helper.registerAllowedMethod('containerPushToRegistry', [Map.class], { m ->
            pushParams = m
            return
        })
        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'docker',
            dockerImageName: 'path/to/myImage',
            dockerRegistryUrl: 'https://my.registry:55555'
        )

        assertThat(pushParams.dockerBuildImage.image.toString(), is('path/to/myImage:1.0.0'))
        assertThat(pushParams.dockerRegistryUrl.toString(), is('https://my.registry:55555'))
        assertThat(nullScript.commonPipelineEnvironment.getValue('containerImage').toString(), is('path/to/myImage:1.0.0'))
    }

    @Test
    void testDockerNoPush() {
        binding.setVariable('docker', new DockerMock('test'))
        def pushParams = [:]
        helper.registerAllowedMethod('containerPushToRegistry', [Map.class], { m ->
            pushParams = m
            return
        })
        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'docker',
            dockerImageName: 'path/to/myImage',
            dockerImageTag: 'myTag',
            dockerRegistryUrl: ''
        )

        assertThat(pushParams.dockerBuildImage, nullValue())
        assertThat(pushParams.dockerRegistryUrl, nullValue())
    }

    @Test
    void testKaniko() {
        def kanikoParams = [:]
        helper.registerAllowedMethod('kanikoExecute', [Map.class], { m ->
            kanikoParams = m
            return
        })

        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'kaniko',
            dockerImageName: 'path/to/myImage',
            dockerImageTag: 'myTag',
            dockerRegistryUrl: 'https://my.registry:55555'
        )

        assertThat(kanikoParams.containerImageNameAndTag.toString(), is('my.registry:55555/path/to/myImage:myTag'))
    }

    @Test
    void testKanikoNoPush() {
        def kanikoParams = [:]
        helper.registerAllowedMethod('kanikoExecute', [Map.class], { m ->
            kanikoParams = m
            return
        })

        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'kaniko',
            dockerImageName: 'path/to/myImage',
            dockerImageTag: 'myTag',
            dockerRegistryUrl: ''
        )

        assertThat(kanikoParams.containerImageNameAndTag, is(''))
    }

    @Test
    void testSwitchToKaniko() {
        shellCallRule.setReturnValue('docker ps -q > /dev/null', 1)
        def kanikoParams = [:]
        helper.registerAllowedMethod('kanikoExecute', [Map.class], { m ->
            kanikoParams = m
            return
        })

        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'kaniko',
            dockerImageName: 'path/to/myImage',
            dockerImageTag: 'myTag',
            dockerRegistryUrl: 'https://my.registry:55555'
        )

        assertThat(kanikoParams.containerImageNameAndTag.toString(), is('my.registry:55555/path/to/myImage:myTag'))
    }

}
