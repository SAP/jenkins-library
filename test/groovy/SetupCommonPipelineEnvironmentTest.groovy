import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import org.yaml.snakeyaml.Yaml

import com.lesfurets.jenkins.unit.BasePipelineTest

import util.Rules

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertNotNull

class SetupCommonPipelineEnvironmentTest extends BasePipelineTest {

    def usedConfigFile

    def setupCommonPipelineEnvironmentScript

    def commonPipelineEnvironment

    @Rule
    public RuleChain rules = Rules.getCommonRules(this)

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

        setupCommonPipelineEnvironmentScript = loadScript("setupCommonPipelineEnvironment.groovy").setupCommonPipelineEnvironment
        commonPipelineEnvironment = loadScript('commonPipelineEnvironment.groovy').commonPipelineEnvironment
    }

    @Test
    void testIsConfigurationAvailable() throws Exception {
        setupCommonPipelineEnvironmentScript.call(script: [commonPipelineEnvironment: commonPipelineEnvironment])

        assertEquals('.pipeline/config.yml', usedConfigFile)
        assertNotNull(commonPipelineEnvironment.configuration)
        assertEquals('develop', commonPipelineEnvironment.configuration.general.productiveBranch)
        assertEquals('my-maven-docker', commonPipelineEnvironment.configuration.steps.mavenExecute.dockerImage)
    }
}
