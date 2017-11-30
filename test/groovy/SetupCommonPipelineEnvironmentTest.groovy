import junit.framework.TestCase
import org.junit.Before
import org.junit.Test
import org.yaml.snakeyaml.Yaml

class SetupCommonPipelineEnvironmentTest extends PiperTestBase {

    def usedConfigFile

    @Before
    void setUp() {
        super.setUp()

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
            return path.endsWith('pipeline_config.yml')
        })
    }

    @Test
    void testIsConfigurationAvailable() throws Exception {
        def script = loadScript("test/resources/pipelines/setupCommonPipelineEnvironmentTest/loadConfiguration.groovy")
            script.execute()

            TestCase.assertEquals('pipeline_config.yml', usedConfigFile)
            TestCase.assertNotNull(script.commonPipelineEnvironment.configuration)
            TestCase.assertEquals('develop', script.commonPipelineEnvironment.configuration.general.productiveBranch)
            TestCase.assertEquals('my-maven-docker', script.commonPipelineEnvironment.configuration.steps.mavenExecute.dockerImage)
    }
}
