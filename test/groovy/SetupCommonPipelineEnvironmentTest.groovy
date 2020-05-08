import com.sap.piper.DefaultValueCache
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import org.yaml.snakeyaml.Yaml
import util.BasePiperTest
import util.JenkinsReadFileRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.hamcrest.Matchers.hasItem
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertNotNull
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue

class SetupCommonPipelineEnvironmentTest extends BasePiperTest {

    def usedConfigFile

    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsReadFileRule readFileRule = new JenkinsReadFileRule(this, "./")

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(stepRule)
        .around(writeFileRule)
        .around(thrown)
        .around(shellRule)
        .around(readFileRule)


    @Before
    void init() {

        def examplePipelineConfig = new File('test/resources/test_pipeline_config.yml').text

        helper.registerAllowedMethod("libraryResource", [String], { fileName ->
            switch(fileName) {
                case 'default_pipeline_environment.yml': return "default: 'config'"
                case 'custom.yml': return "custom: 'myConfig'"
                case 'notFound.yml': throw new hudson.AbortException('No such library resource notFound could be found')
                default: return "the:'end'"
            }
        })

        helper.registerAllowedMethod("readYaml", [Map], { Map parameters ->
            Yaml yamlParser = new Yaml()
            if (parameters.text) {
                return yamlParser.load(parameters.text)
            } else if(parameters.file) {
                if(parameters.file == '.pipeline/default_pipeline_environment.yml') return [default: 'config']
                else if (parameters.file == '.pipeline/custom.yml') return [custom: 'myConfig']
            } else {
                throw new IllegalArgumentException("Key 'text' and 'file' are both missing in map ${m}.")
            }
            usedConfigFile = parameters.file
            return yamlParser.load(examplePipelineConfig)
        })
    }

    @Test
    void testIsYamlConfigurationAvailable() throws Exception {

        helper.registerAllowedMethod("fileExists", [String], { String path ->
            return path.endsWith('.pipeline/config.yml')
        })

        stepRule.step.setupCommonPipelineEnvironment(script: nullScript)

        assertEquals('.pipeline/config.yml', usedConfigFile)
        assertNotNull(nullScript.commonPipelineEnvironment.configuration)
        assertEquals('develop', nullScript.commonPipelineEnvironment.configuration.general.productiveBranch)
        assertEquals('my-maven-docker', nullScript.commonPipelineEnvironment.configuration.steps.mavenExecute.dockerImage)
    }

    @Test
    void testWorksAlsoWithYamlFileEnding() throws Exception {

        helper.registerAllowedMethod("fileExists", [String], { String path ->
            return path.endsWith('.pipeline/config.yaml')
        })

        stepRule.step.setupCommonPipelineEnvironment(script: nullScript)

        assertEquals('.pipeline/config.yaml', usedConfigFile)
        assertNotNull(nullScript.commonPipelineEnvironment.configuration)
        assertEquals('develop', nullScript.commonPipelineEnvironment.configuration.general.productiveBranch)
        assertEquals('my-maven-docker', nullScript.commonPipelineEnvironment.configuration.steps.mavenExecute.dockerImage)
    }

    @Test
    public void testAttemptToLoadNonExistingConfigFile() {

        helper.registerAllowedMethod("fileExists", [String], { String path ->
            switch(path) {
                case 'default_pipeline_environment.yml': return false
                case 'custom.yml': return false
                case 'notFound.yml': return false
                default: return true
            }
        })

        helper.registerAllowedMethod("handlePipelineStepErrors", [Map,Closure], { Map map, Closure closure ->
            closure()
        })

        // Behavior documented here based on reality check
        thrown.expect(hudson.AbortException.class)
        thrown.expectMessage('No such library resource notFound could be found')

        stepRule.step.setupCommonPipelineEnvironment(
            script: nullScript,
            customDefaults: 'notFound.yml'
        )
    }

    @Test
    void testDefaultPipelineEnvironmentWithCustomConfigReferencedAsString() {
        helper.registerAllowedMethod("prepareDefaultValues", [Map], {Map parameters ->
            assertTrue(parameters.customDefaults instanceof List)
        })

        helper.registerAllowedMethod("fileExists", [String], {String path ->
            switch (path) {
                case 'default_pipeline_environment.yml': return false
                case 'custom.yml': return false
                default: return true
            }
        })

        stepRule.step.setupCommonPipelineEnvironment(
            script: nullScript,
            customDefaults: 'custom.yml'
        )
    }

    @Test
    void testAttemptToLoadFileFromURL() {
        helper.registerAllowedMethod("prepareDefaultValues", [Map], {Map parameters ->
            assertTrue(parameters.customDefaultsFromConfig instanceof List)
            assertTrue(parameters.customDefaultsFromConfig.contains('.pipeline/custom_default_from_url_0.yml'))
        })

        helper.registerAllowedMethod("fileExists", [String], {String path ->
            switch (path) {
                case 'default_pipeline_environment.yml': return false
                default: return true
            }
        })

        String customDefaultUrl = "https://url-to-my-config.com/my-config.yml"

        helper.registerAllowedMethod("readYaml", [Map], { Map parameters ->
            Yaml yamlParser = new Yaml()
            if (parameters.text) {
                return yamlParser.load(parameters.text)
            } else if (parameters.file) {
                if (parameters.file == '.pipeline/config-with-custom-defaults.yml') {
                    return [customDefaults: "${customDefaultUrl}"]
                }
            }
            throw new IllegalArgumentException("Unexpected invocation of readYaml step")
        })

        stepRule.step.setupCommonPipelineEnvironment(
            script: nullScript,
            configFile: '.pipeline/config-with-custom-defaults.yml',
        )

        assertThat(shellRule.shell, hasItem("curl --fail --location --output .pipeline/custom_default_from_url_0.yml " + customDefaultUrl))
    }

    @Test
    void testAttemptToLoadFileFromWorkspace() {
        String customDefaultPath = "./my-config.yml"
        new File(customDefaultPath).write('custom: \'myConfig\'')
        helper.registerAllowedMethod("prepareDefaultValues", [Map], {Map parameters ->
            assertTrue(parameters.customDefaults instanceof List)
            assertTrue(parameters.customDefaults.contains(customDefaultPath))
        })

        helper.registerAllowedMethod("fileExists", [String], {String path ->
            switch (path) {
                case 'default_pipeline_environment.yml': return false
                case customDefaultPath: return true
                default: return true
            }
        })

        stepRule.step.setupCommonPipelineEnvironment(
            script: nullScript,
            customDefaults: customDefaultPath
        )
    }
}

