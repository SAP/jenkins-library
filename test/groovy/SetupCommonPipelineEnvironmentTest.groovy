import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import org.yaml.snakeyaml.Yaml

import com.sap.piper.Utils

import util.BasePiperTest
import util.Rules
import util.JenkinsReadYamlRule
import util.JenkinsStepRule

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertNotNull


class SetupCommonPipelineEnvironmentTest extends BasePiperTest {
    def usedConfigFile
    def swaOldConfigUsed

    private JenkinsStepRule jsr = new JenkinsStepRule(this)

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
        helper.registerAllowedMethod("readProperties", [Map], { Map parameters ->
            usedConfigFile = parameters.file
            Properties props = new Properties()
            props.setProperty('key', 'value')
            return props
        })

        swaOldConfigUsed = null
    }

    @Test
    void testIsYamlConfigurationAvailable() throws Exception {

        helper.registerAllowedMethod("fileExists", [String], { String path ->
            return path.endsWith('.pipeline/config.yml')
        })

        jsr.step.setupCommonPipelineEnvironment(script: nullScript, utils: getSWAMockedUtils())

        assertEquals(Boolean.FALSE.toString(), swaOldConfigUsed)
        assertEquals('.pipeline/config.yml', usedConfigFile)
        assertNotNull(nullScript.commonPipelineEnvironment.configuration)
        assertEquals('develop', nullScript.commonPipelineEnvironment.configuration.general.productiveBranch)
        assertEquals('my-maven-docker', nullScript.commonPipelineEnvironment.configuration.steps.mavenExecute.dockerImage)
    }

    @Test
    void testIsPropertiesConfigurationAvailable() {

        helper.registerAllowedMethod("fileExists", [String], { String path ->
            return path.endsWith('.pipeline/config.properties')
        })

        jsr.step.setupCommonPipelineEnvironment(script: nullScript, utils: getSWAMockedUtils())

        assertEquals(Boolean.TRUE.toString(), swaOldConfigUsed)
        assertEquals('.pipeline/config.properties', usedConfigFile)
        assertNotNull(nullScript.commonPipelineEnvironment.configProperties)
        assertEquals('value', nullScript.commonPipelineEnvironment.configProperties['key'])
    }

    private getSWAMockedUtils() {
        new Utils() {
            void pushToSWA(Map payload, Map config) {
                SetupCommonPipelineEnvironmentTest.this.swaOldConfigUsed = payload.stepParam5
            }
        }
    }
}
