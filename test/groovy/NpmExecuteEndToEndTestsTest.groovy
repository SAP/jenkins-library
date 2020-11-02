import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsMockStepRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

import com.sap.piper.Utils

import static org.junit.Assert.assertFalse
import static org.junit.Assert.assertTrue
import static org.junit.Assert.assertTrue

class NpmExecuteEndToEndTestsTest extends BasePiperTest {

    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsMockStepRule npmExecuteScriptsRule = new JenkinsMockStepRule(this, 'npmExecuteScripts')
    private JenkinsCredentialsRule credentialsRule = new JenkinsCredentialsRule(this)
    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this)

    private boolean executedOnKubernetes = false
    private boolean executedOnNode = false
    private boolean executedInParallel = false

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(thrown)
        .around(readYamlRule)
        .around(credentialsRule)
        .around(stepRule)
        .around(npmExecuteScriptsRule)

    @Before
    void init() {
        helper.registerAllowedMethod("deleteDir", [], null)

        helper.registerAllowedMethod('dockerExecuteOnKubernetes', [Map.class, Closure.class], {params, body ->
            executedOnKubernetes = true
            body()
        })
        helper.registerAllowedMethod('node', [String.class, Closure.class], {s, body ->
            executedOnNode = true
            body()
        })
        helper.registerAllowedMethod("parallel", [Map.class], { map ->
            map.each {key, value ->
                value()
            }
            executedInParallel = true
        })
        helper.registerAllowedMethod('findFiles', [Map.class], {return []})

        credentialsRule.reset()
            .withCredentials('testCred', 'test_cf', '********')
            .withCredentials('testCred2', 'test_other', '**')

        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
    }

    @Test
    void noAppUrl() {
        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[npmExecuteEndToEndTests] The execution failed, since no appUrls are defined. Please provide appUrls as a list of maps.')

        stepRule.step.npmExecuteEndToEndTests(
            script: nullScript,
            stageName: "myStage",
            runScript: "ci-e2e"
        )
    }

    @Test
    void noRunScript() {
        def appUrl = [url: "http://my-url.com"]

        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            appUrls: [appUrl]
        ]]]

        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[npmExecuteEndToEndTests] No runScript was defined.')

        stepRule.step.npmExecuteEndToEndTests(
            script: nullScript,
            stageName: "myStage"
        )
    }

    @Test
    void appUrlsNoList() {
        def appUrl = "http://my-url.com"

        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            appUrls: appUrl
        ]]]

        thrown.expect(hudson.AbortException)
        thrown.expectMessage("[npmExecuteEndToEndTests] The execution failed, since appUrls is not a list. Please provide appUrls as a list of maps.")

        stepRule.step.npmExecuteEndToEndTests(
            script: nullScript,
            stageName: "myStage",
            runScript: "ci-e2e"
        )
    }

    @Test
    void appUrlsNoMap() {
        def appUrl = "http://my-url.com"

        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            appUrls: [appUrl]
        ]]]

        thrown.expect(hudson.AbortException)
        thrown.expectMessage("[npmExecuteEndToEndTests] The element ${appUrl} is not of type map. Please provide appUrls as a list of maps.")

        stepRule.step.npmExecuteEndToEndTests(
            script: nullScript,
            stageName: "myStage",
            runScript: "ci-e2e"
        )
    }

    @Test
    void appUrlParametersNoList() {
        def appUrl = [url: "http://my-url.com", credentialId: 'testCred', parameters: '--tag scenario1']

        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            appUrls: [appUrl]
        ]]]

        thrown.expect(hudson.AbortException)
        thrown.expectMessage("[npmExecuteEndToEndTests] The parameters property is not of type list. Please provide parameters as a list of strings.")

        stepRule.step.npmExecuteEndToEndTests(
            script: nullScript,
            stageName: "myStage",
            runScript: "ci-e2e"
        )
    }

    @Test
    void oneAppUrl() {
        def appUrl = [url: "http://my-url.com"]

        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            appUrls: [appUrl]
        ]]]

        stepRule.step.npmExecuteEndToEndTests(
            script: nullScript,
            stageName: "myStage",
            runScript: "ci-e2e"
        )

        assertFalse(executedInParallel)
        assert npmExecuteScriptsRule.hasParameter('script', nullScript)
        assert npmExecuteScriptsRule.hasParameter('parameters', [dockerOptions: ['--shm-size 512MB']])
        assert npmExecuteScriptsRule.hasParameter('virtualFrameBuffer', true)
        assert npmExecuteScriptsRule.hasParameter('runScripts', ["ci-e2e"])
        assert npmExecuteScriptsRule.hasParameter('scriptOptions', ["--launchUrl=${appUrl.url}"])
    }

    @Test
    void oneAppUrlWithCredentials() {
        def appUrl = [url: "http://my-url.com", credentialId: 'testCred']

        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            appUrls: [appUrl]
        ]]]

        stepRule.step.npmExecuteEndToEndTests(
            script: nullScript,
            stageName: "myStage",
            runScript: "ci-e2e"
        )

        assert npmExecuteScriptsRule.hasParameter('script', nullScript)
        assert npmExecuteScriptsRule.hasParameter('parameters', [dockerOptions: ['--shm-size 512MB']])
        assert npmExecuteScriptsRule.hasParameter('virtualFrameBuffer', true)
        assert npmExecuteScriptsRule.hasParameter('runScripts', ["ci-e2e"])
        assert npmExecuteScriptsRule.hasParameter('scriptOptions', ["--launchUrl=${appUrl.url}"])
    }

    @Test
    void twoAppUrlsWithCredentials() {
        def appUrl = [url: "http://my-url.com", credentialId: 'testCred']
        def appUrl2 = [url: "http://my-second-url.com", credentialId: 'testCred2']

        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            appUrls: [appUrl, appUrl2]
        ]]]

        stepRule.step.npmExecuteEndToEndTests(
            script: nullScript,
            stageName: "myStage",
            runScript: "ci-e2e"
        )

        assert npmExecuteScriptsRule.hasParameter('script', nullScript)
        assert npmExecuteScriptsRule.hasParameter('parameters', [dockerOptions: ['--shm-size 512MB']])
        assert npmExecuteScriptsRule.hasParameter('virtualFrameBuffer', true)
        assert npmExecuteScriptsRule.hasParameter('runScripts', ["ci-e2e"])
        assert npmExecuteScriptsRule.hasParameter('scriptOptions', ["--launchUrl=${appUrl.url}"])
        assert npmExecuteScriptsRule.hasParameter('scriptOptions', ["--launchUrl=${appUrl2.url}"])
    }

    @Test
    void oneAppUrlWithCredentialsAndParameters() {
        def appUrl = [url: "http://my-url.com", credentialId: 'testCred', parameters: ['--tag','scenario1', '--NIGHTWATCH_ENV=chrome']]

        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            appUrls: [appUrl]
        ]]]

        stepRule.step.npmExecuteEndToEndTests(
            script: nullScript,
            stageName: "myStage",
            buildDescriptorExcludeList: ["path/to/package.json"],
            runScript: "ci-e2e"
        )

        assert npmExecuteScriptsRule.hasParameter('script', nullScript)
        assert npmExecuteScriptsRule.hasParameter('parameters', [dockerOptions: ['--shm-size 512MB']])
        assert npmExecuteScriptsRule.hasParameter('virtualFrameBuffer', true)
        assert npmExecuteScriptsRule.hasParameter('runScripts', ["ci-e2e"])
        assert npmExecuteScriptsRule.hasParameter('scriptOptions', ["--launchUrl=${appUrl.url}"] + appUrl.parameters)
        assert npmExecuteScriptsRule.hasParameter('buildDescriptorExcludeList', ["path/to/package.json"])
    }

    @Test
    void parallelE2eTest() {
        def appUrl = [url: "http://my-url.com", credentialId: 'testCred']

        nullScript.commonPipelineEnvironment.configuration = [
            general: [parallelExecution: true],
            stages: [
                myStage:[
                    appUrls: [appUrl]
                ]]]

        stepRule.step.npmExecuteEndToEndTests(
            script: nullScript,
            stageName: "myStage",
            runScript: "ci-e2e"
        )

        assertTrue(executedInParallel)
        assertTrue(executedOnNode)
        assertFalse(executedOnKubernetes)
    }

    @Test
    void parallelE2eTestOnKubernetes() {
        def appUrl = [url: "http://my-url.com", credentialId: 'testCred']
        binding.variables.env.POD_NAME = "name"

        nullScript.commonPipelineEnvironment.configuration = [
            general: [parallelExecution: true],
            stages: [
                myStage:[
                    appUrls: [appUrl]
                ]]]

        stepRule.step.npmExecuteEndToEndTests(
            script: nullScript,
            stageName: "myStage",
            runScript: "ci-e2e"
        )

        assertTrue(executedInParallel)
        assertFalse(executedOnNode)
        assertTrue(executedOnKubernetes)
    }
}
