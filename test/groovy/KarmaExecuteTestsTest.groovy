import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class KarmaExecuteTestsTest extends BasePiperTest {
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsEnvironmentRule environmentRule = new JenkinsEnvironmentRule(this)
    private ExpectedException thrown = ExpectedException.none()

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(shellRule)
        .around(loggingRule)
        .around(environmentRule)
        .around(stepRule)
        .around(thrown)

    def seleniumParams = [:]

    @Before
    void init() throws Exception {
        helper.registerAllowedMethod("unstash", [String.class], { s -> return [s]})

        helper.registerAllowedMethod('seleniumExecuteTests', [Map.class, Closure.class], {map, body ->
            seleniumParams = map
            return body()
        })
    }

    @Test
    void testDefaults() throws Exception {
        stepRule.step.karmaExecuteTests(
            script: nullScript,
            juStabUtils: utils
        )
        assertThat(shellRule.shell, hasItems(
            containsString("cd '.' && npm install --quiet"),
            containsString("cd '.' && npm run karma")
        ))
        assertThat(seleniumParams.dockerImage, is('node:lts-stretch'))
        assertThat(seleniumParams.dockerName, is('karma'))
        assertThat(seleniumParams.dockerWorkspace, is('/home/node'))
        assertJobStatusSuccess()
    }

    @Test
    void testDockerFromCustomStepConfiguration() {

        def expectedImage = 'image:test'
        def expectedEnvVars = ['NO_PROXY':'', 'no_proxy':'', 'env1': 'value1', 'env2': 'value2']
        def expectedOptions = '--opt1=val1 --opt2=val2 --opt3'
        def expectedWorkspace = '/path/to/workspace'
        
        nullScript.commonPipelineEnvironment.configuration = [steps:[karmaExecuteTests:[
            dockerImage: expectedImage, 
            dockerOptions: expectedOptions,
            dockerEnvVars: expectedEnvVars,
            dockerWorkspace: expectedWorkspace
            ]]]

        stepRule.step.karmaExecuteTests(
            script: nullScript,
            juStabUtils: utils
        )
        
        assert expectedImage == seleniumParams.dockerImage
        assert expectedOptions == seleniumParams.dockerOptions
        assert expectedWorkspace == seleniumParams.dockerWorkspace
        expectedEnvVars.each { key, value ->
            assert seleniumParams.dockerEnvVars[key] == value
        }
    }

    @Test
    void testMultiModules() throws Exception {
        stepRule.step.karmaExecuteTests(
            script: nullScript,
            juStabUtils: utils,
            modules: ['./ui-trade', './ui-traderequest']
        )
        assertThat(shellRule.shell, hasItems(
            containsString("cd './ui-trade' && npm run karma"),
            containsString("cd './ui-trade' && npm install --quiet")
        ))
        assertThat(shellRule.shell, hasItems(
            containsString("cd './ui-traderequest' && npm run karma"),
            containsString("cd './ui-traderequest' && npm install --quiet")
        ))
        assertJobStatusSuccess()
    }
}
