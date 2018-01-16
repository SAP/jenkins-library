import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.yaml.snakeyaml.Yaml

import util.JenkinsSetupRule

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertNotNull

class SetupCommonPipelineEnvironmentTest extends PiperTestBase {

    def usedConfigFile

    @Rule
    public JenkinsSetupRule jsr = new JenkinsSetupRule(this)

    @Before
    void init() {

        def examplePipelineConfig = new File('test/resources/test_pipeline_config.yml').text

        helper.registerAllowedMethod("readYaml", [Map], { Map parameters ->
            Yaml yamlParser = new Yaml()
            if(parameters.text) {
                return yamlParser.load(parameters.text)
            }

            usedConfigFile = parameters.file
            return yamlParser.load(examplePipelineConfig)
        })

        helper.registerAllowedMethod("fileExists", [String], { String path ->
            return path.endsWith('.pipeline/config.yml')
        })
    }

    @Test
    void testIsConfigurationAvailable() throws Exception {
        def script = loadScript("test/resources/pipelines/setupCommonPipelineEnvironmentTest/loadConfiguration.groovy")
        script.execute()

        assertEquals('.pipeline/config.yml', usedConfigFile)
        assertNotNull(script.commonPipelineEnvironment.configuration)
        assertEquals('develop', script.commonPipelineEnvironment.configuration.general.productiveBranch)
        assertEquals('my-maven-docker', script.commonPipelineEnvironment.configuration.steps.mavenExecute.dockerImage)
    }
}
