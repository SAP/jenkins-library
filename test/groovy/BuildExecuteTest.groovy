#!groovy
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

import static org.hamcrest.CoreMatchers.containsString
import static org.hamcrest.CoreMatchers.hasItem
import static org.hamcrest.CoreMatchers.is
import static org.hamcrest.CoreMatchers.nullValue
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertFalse
import static org.junit.Assert.assertNotNull
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue
import static org.junit.Assert.fail

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
        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
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
    void inferBuildToolMaven() {
        boolean buildToolCalled = false
        helper.registerAllowedMethod('fileExists', [String.class], { s ->
            return s == "pom.xml"
        })
        helper.registerAllowedMethod('mavenBuild', [Map.class], { m ->
            buildToolCalled = true
            return
        })

        setupCommonPipelineEnvironment.inferBuildTool(nullScript, [inferBuildTool: true])
        stepRule.step.buildExecute(
            script: nullScript,
        )

        assertNotNull(nullScript.commonPipelineEnvironment.getBuildTool())
        assertEquals('maven', nullScript.commonPipelineEnvironment.getBuildTool())
        assertTrue(buildToolCalled)
    }

    @Test
    void inferBuildToolNpm() {
        boolean buildToolCalled = false
        helper.registerAllowedMethod('fileExists', [String.class], { s ->
            return s == "package.json"
        })
        helper.registerAllowedMethod('npmExecuteScripts', [Map.class], { m ->
            buildToolCalled = true
            return
        })

        setupCommonPipelineEnvironment.inferBuildTool(nullScript, [inferBuildTool: true])
        stepRule.step.buildExecute(
            script: nullScript,
        )

        assertNotNull(nullScript.commonPipelineEnvironment.getBuildTool())
        assertEquals('npm', nullScript.commonPipelineEnvironment.getBuildTool())
        assertTrue(buildToolCalled)
    }

    @Test
    void inferBuildToolMTA() {
        boolean buildToolCalled = false
        helper.registerAllowedMethod('fileExists', [String.class], { s ->
            return s == "mta.yaml"
        })
        helper.registerAllowedMethod('mtaBuild', [Map.class], { m ->
            buildToolCalled = true
            return
        })

        setupCommonPipelineEnvironment.inferBuildTool(nullScript, [inferBuildTool: true])
        stepRule.step.buildExecute(
            script: nullScript,
        )

        assertNotNull(nullScript.commonPipelineEnvironment.getBuildTool())
        assertEquals('mta', nullScript.commonPipelineEnvironment.getBuildTool())
        assertTrue(buildToolCalled)
    }

    @Test
    void 'Do not infer build tool, do not set build tool, with docker dockerImage and dockerCommand, should run docker'() {
        helper.registerAllowedMethod('fileExists', [String.class], { s ->
            return s == "package.json"
        })
        helper.registerAllowedMethod('npmExecuteScripts', [Map.class], { m ->
            fail("Called npmExecuteScripts which should not happen when no buildTool was defined but dockerImage and dockerCommand were.")
        })

        // Does nothing because feature toggle is not active
        setupCommonPipelineEnvironment.inferBuildTool(nullScript, [inferBuildTool: false])

        stepRule.step.buildExecute(
            script: nullScript,
            dockerImage: 'path/to/myImage:tag',
            dockerCommand: 'myTestCommand'
        )

        assertThat(dockerRule.dockerParams.dockerImage, is('path/to/myImage:tag'))
        assertThat(shellCallRule.shell, hasItem('myTestCommand'))
    }

    @Test
    void 'Do infer build tool, do not set build tool, with docker dockerImage and dockerCommand, should run npm'() {
        boolean npmCalled = false
        helper.registerAllowedMethod('fileExists', [String.class], { s ->
            return s == "package.json"
        })
        helper.registerAllowedMethod('npmExecuteScripts', [Map.class], { m ->
            npmCalled = true
            return
        })

        setupCommonPipelineEnvironment.inferBuildTool(nullScript, [inferBuildTool: true])

        stepRule.step.buildExecute(
            script: nullScript,
            dockerImage: 'path/to/myImage:tag',
            dockerCommand: 'myTestCommand'
        )

        assertTrue(npmCalled)
        assertEquals(0, shellCallRule.shell.size())
    }

    @Test
    void testMaven() {
        boolean buildToolCalled = false
        boolean installOptionSet = false
        helper.registerAllowedMethod('mavenBuild', [Map.class], { m ->
            buildToolCalled = true
            return
        })
        helper.registerAllowedMethod('npmExecuteScripts', [Map.class], { m ->
            installOptionSet = m['install']
            return
        })
        helper.registerAllowedMethod('fileExists', [String.class], { s ->
            return s == 'package.json'
        })
        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'maven',
        )
        assertTrue(buildToolCalled)
        assertTrue(installOptionSet)
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
        boolean buildToolCalled = false
        binding.setVariable('docker', new DockerMock('test'))
        def pushParams = [:]
        helper.registerAllowedMethod('kanikoExecute', [Map.class], { m ->
            buildToolCalled = true
            return
        })
        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'docker',
        )
        assertThat(buildToolCalled, is(true))
    }

    @Test
    void testDockerWithoutCNB() {
        boolean kanikoExecuteCalled = false
        boolean cnbBuildCalled = false
        binding.setVariable('docker', new DockerMock('test'))
        def pushParams = [:]
        helper.registerAllowedMethod('kanikoExecute', [Map.class], { m ->
            kanikoExecuteCalled = true
            return
        })
        helper.registerAllowedMethod('cnbBuild', [Map.class], { m ->
            cnbBuildCalled = true
            return
        })
        stepRule.step.buildExecute(
                script: nullScript,
                buildTool: 'docker',
        )
        assertThat(cnbBuildCalled, is(false))
        assertThat(kanikoExecuteCalled, is(true))
    }

    @Test
    void testDockerWithCNB() {
        boolean kanikoExecuteCalled = false
        boolean cnbBuildCalled = false
        binding.setVariable('docker', new DockerMock('test'))
        def pushParams = [:]
        helper.registerAllowedMethod('kanikoExecute', [Map.class], { m ->
            kanikoExecuteCalled = true
            return
        })
        helper.registerAllowedMethod('cnbBuild', [Map.class], { m ->
            cnbBuildCalled = true
            return
        })
        stepRule.step.buildExecute(
                script: nullScript,
                buildTool: 'docker',
                cnbBuild: true
        )
        assertThat(cnbBuildCalled, is(true))
        assertThat(kanikoExecuteCalled, is(false))
    }

    @Test
    void testKaniko() {
        binding.setVariable('docker', new DockerMock('test'))
        def buildToolCalled = false
        helper.registerAllowedMethod('kanikoExecute', [Map.class], { m ->
            buildToolCalled = true
            return
        })
        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'kaniko',
        )
        assertThat(buildToolCalled, is(true))
    }

    @Test
    void testCnbBuildCalledWhenConfigured() {
        def cnbBuildCalled = false
        def npmExecuteScriptsCalled = false
        helper.registerAllowedMethod('npmExecuteScripts', [Map.class], { m ->
            npmExecuteScriptsCalled = true
        })
        helper.registerAllowedMethod('cnbBuild', [Map.class], { m ->
            cnbBuildCalled = true
            return
        })
        assertThat(nullScript.commonPipelineEnvironment.getContainerProperty('buildpacks'), nullValue())

        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'npm',
            cnbBuild: true
        )

        assertThat(npmExecuteScriptsCalled, is(true))
        assertThat(cnbBuildCalled, is(true))
    }

    @Test
    void testCnbBuildNotCalledWhenNotConfigured() {
        def cnbBuildCalled = false
        def npmExecuteScriptsCalled = false
        helper.registerAllowedMethod('npmExecuteScripts', [Map.class], { m ->
            npmExecuteScriptsCalled = true
        })
        helper.registerAllowedMethod('cnbBuild', [Map.class], { m ->
            cnbBuildCalled = true
            return
        })
        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'npm',
            cnbBuild: false
        )
        assertThat(npmExecuteScriptsCalled, is(true))
        assertThat(cnbBuildCalled, is(false))
    }

    @Test
    void testHelmExecuteCalledWhenConfigured() {
        def helmExecuteCalled = false
        helper.registerAllowedMethod('helmExecute', [Map.class], { m ->
            helmExecuteCalled = true
            return
        })
        helper.registerAllowedMethod('npmExecuteScripts', [Map.class], { m ->
        })

        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'npm',
            helmExecute: true
        )

        assertThat(helmExecuteCalled, is(true))
    }

    @Test
    void testHelmExecuteNotCalledWhenNotConfigured() {
        def helmExecuteCalled = false
        helper.registerAllowedMethod('helmExecute', [Map.class], { m ->
            helmExecuteCalled = true
            return
        })
        helper.registerAllowedMethod('npmExecuteScripts', [Map.class], { m ->
        })
        stepRule.step.buildExecute(
            script: nullScript,
            buildTool: 'npm',
            helmExecute: false
        )

        assertThat(helmExecuteCalled, is(false))
    }
}
