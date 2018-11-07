import org.junit.Before
import org.junit.Rule;
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain;
import com.sap.piper.DefaultValueCache

import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule;

import util.Rules

public class PrepareDefaultValuesTest extends BasePiperTest {

    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(jsr)
        .around(jlr)

    @Before
    public void setup() {

        helper.registerAllowedMethod("libraryResource", [String], { fileName ->
            switch(fileName) {
                case 'default_pipeline_environment.yml': return "default: 'config'"
                case 'custom.yml': return "custom: 'myConfig'"
                case 'not_found': throw new hudson.AbortException('No such library resource not_found could be found')
                default: return "the:'end'"
            }
        })
    }

    @Test
    public void testDefaultPipelineEnvironmentOnly() {

        jsr.step.prepareDefaultValue(script: nullScript)

        assert DefaultValueCache.getInstance().getDefaultValues().size() == 1
        assert DefaultValueCache.getInstance().getDefaultValues().default == 'config'
    }

    @Test
    public void testReInitializeOnCustomConfig() {

        def instance = DefaultValueCache.createInstance([key:'value'])

        // existing instance is dropped in case a custom config is provided.
        jsr.step.prepareDefaultValue(script: nullScript, customDefaults: 'custom.yml')

        // this check is for checking we have another instance
        assert ! instance.is(DefaultValueCache.getInstance())

        // some additional checks that the configuration represented by the new
        // config is fine
        assert DefaultValueCache.getInstance().getDefaultValues().size() == 2
        assert DefaultValueCache.getInstance().getDefaultValues().default == 'config'
        assert DefaultValueCache.getInstance().getDefaultValues().custom == 'myConfig'
    }

    @Test
    public void testNoReInitializeWithoutCustomConfig() {

        def instance = DefaultValueCache.createInstance([key:'value'])

        jsr.step.prepareDefaultValue(script: nullScript)

        assert instance.is(DefaultValueCache.getInstance())
        assert DefaultValueCache.getInstance().getDefaultValues().size() == 1
        assert DefaultValueCache.getInstance().getDefaultValues().key == 'value'
    }

    @Test
    public void testAttemptToLoadNonExistingConfigFile() {

        // Behavior documented here based on reality check
        thrown.expect(hudson.AbortException.class)
        thrown.expectMessage('No such library resource not_found could be found')

        jsr.step.prepareDefaultValue(script: nullScript, customDefaults: 'not_found')
    }

    @Test
    public void testDefaultPipelineEnvironmentWithCustomConfigReferencedAsString() {

        jsr.step.prepareDefaultValue(script: nullScript, customDefaults: 'custom.yml')

        assert DefaultValueCache.getInstance().getDefaultValues().size() == 2
        assert DefaultValueCache.getInstance().getDefaultValues().default == 'config'
        assert DefaultValueCache.getInstance().getDefaultValues().custom == 'myConfig'
    }

    @Test
    public void testDefaultPipelineEnvironmentWithCustomConfigReferencedAsList() {

        jsr.step.prepareDefaultValue(script: nullScript, customDefaults: ['custom.yml'])

        assert DefaultValueCache.getInstance().getDefaultValues().size() == 2
        assert DefaultValueCache.getInstance().getDefaultValues().default == 'config'
        assert DefaultValueCache.getInstance().getDefaultValues().custom == 'myConfig'
    }

    @Test
    public void testAssertNoLogMessageInCaseOfNoAdditionalConfigFiles() {

        jsr.step.prepareDefaultValue(script: nullScript)

        assert ! jlr.log.contains("Loading configuration file 'default_pipeline_environment.yml'")
    }

    @Test
    public void testAssertLogMessageInCaseOfMoreThanOneConfigFile() {

        jsr.step.prepareDefaultValue(script: nullScript, customDefaults: ['custom.yml'])

        assert jlr.log.contains("Loading configuration file 'default_pipeline_environment.yml'")
        assert jlr.log.contains("Loading configuration file 'custom.yml'")
    }
}
