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
    void appUrlsNoList() {
        def appUrl = "http://my-url.com"

        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            appUrls: appUrl
        ]]]

        thrown.expect(hudson.AbortException)
        thrown.expectMessage("[npmExecuteEndToEndTests] The execution failed, since appUrls is not a list. Please provide appUrls as a list of maps.")

        stepRule.step.npmExecuteEndToEndTests(
            script: nullScript,
            stageName: "myStage"
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
            stageName: "myStage"
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
            stageName: "myStage"
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
            stageName: "myStage"
        )

        assertFalse(executedInParallel)
        assert npmExecuteScriptsRule.hasParameter('script', nullScript)
        assert npmExecuteScriptsRule.hasParameter('parameters', [dockerOptions: ['--shm-size 512MB']])
        assert npmExecuteScriptsRule.hasParameter('virtualFrameBuffer', true)
        assert npmExecuteScriptsRule.hasParameter('runScripts', ["ci-e2e"])
        assert npmExecuteScriptsRule.hasParameter('scriptOptions', ["--launchUrl=${appUrl.url}"])
    }

    @Test
    void baseUrl() {

        nullScript.commonPipelineEnvironment.configuration = [
                stages: [
                        myStage: [
                            baseUrl: "http://my-url.com"
                        ]
                ]
        ]

        stepRule.step.npmExecuteEndToEndTests(
                script: nullScript,
                stageName: "myStage"
        )

        assertFalse(executedInParallel)
        assert npmExecuteScriptsRule.hasParameter('script', nullScript)
        assert npmExecuteScriptsRule.hasParameter('parameters', [dockerOptions: ['--shm-size 512MB']])
        assert npmExecuteScriptsRule.hasParameter('virtualFrameBuffer', true)
        assert npmExecuteScriptsRule.hasParameter('runScripts', ["ci-e2e"])
        assert npmExecuteScriptsRule.hasParameter('scriptOptions', ["--baseUrl=http://my-url.com"])
    }

    @Test
    void oneAppUrl__whenWdi5IsTrue__wdi5CredentialIsProvided() {
        def appUrl = [url: "http://my-url.com", credentialId: "testCred"]

        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            appUrls: [appUrl],
            wdi5: true
        ]]]

        stepRule.step.npmExecuteEndToEndTests(
            script: nullScript,
            stageName: "myStage"
        )

        assertFalse(executedInParallel)
        assert npmExecuteScriptsRule.hasParameter('script', nullScript)
        assert npmExecuteScriptsRule.hasParameter('parameters', [dockerOptions: ['--shm-size 512MB']])
        assert npmExecuteScriptsRule.hasParameter('virtualFrameBuffer', true)
        assert npmExecuteScriptsRule.hasParameter('runScripts', ["ci-e2e"])
        assert npmExecuteScriptsRule.hasParameter('scriptOptions', ["--launchUrl=${appUrl.url}"])
        assert binding.hasVariable('e2e_username')
        assert binding.hasVariable('e2e_password')
        assert binding.hasVariable('wdi5_username')
        assert binding.hasVariable('wdi5_password')
    }

    @Test
    void oneAppUrl__whenWdi5IsNotSet__noWdi5CredentialIsProvided() {
        def appUrl = [url: "http://my-url.com", credentialId: "testCred"]

        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            appUrls: [appUrl]
        ]]]

        stepRule.step.npmExecuteEndToEndTests(
            script: nullScript,
            stageName: "myStage"
        )

        assertFalse(executedInParallel)
        assert npmExecuteScriptsRule.hasParameter('script', nullScript)
        assert npmExecuteScriptsRule.hasParameter('parameters', [dockerOptions: ['--shm-size 512MB']])
        assert npmExecuteScriptsRule.hasParameter('virtualFrameBuffer', true)
        assert npmExecuteScriptsRule.hasParameter('runScripts', ["ci-e2e"])
        assert npmExecuteScriptsRule.hasParameter('scriptOptions', ["--launchUrl=${appUrl.url}"])
        assert binding.hasVariable('e2e_username')
        assert binding.hasVariable('e2e_password')
        assertFalse binding.hasVariable('wdi5_username')
        assertFalse binding.hasVariable('wdi5_password')
    }

    @Test
    void whenWdi5IsTrue__wdi5CredentialIsProvided() {

        nullScript.commonPipelineEnvironment.configuration = [
                stages: [
                        myStage: [
                            wdi5: true,
                            credentialsId: "testCred"
                        ]
                ]
        ]

        stepRule.step.npmExecuteEndToEndTests(
                script: nullScript,
                stageName: "myStage"
        )

        assertFalse(executedInParallel)
        assert npmExecuteScriptsRule.hasParameter('script', nullScript)
        assert npmExecuteScriptsRule.hasParameter('parameters', [dockerOptions: ['--shm-size 512MB']])
        assert npmExecuteScriptsRule.hasParameter('virtualFrameBuffer', true)
        assert npmExecuteScriptsRule.hasParameter('runScripts', ["ci-e2e"])
        assert binding.hasVariable('e2e_username')
        assert binding.hasVariable('e2e_password')
        assert binding.hasVariable('wdi5_username')
        assert binding.hasVariable('wdi5_password')

    }

    @Test
    void whenWdi5IsNotSet__noWdi5CredentialIsProvided() {

        nullScript.commonPipelineEnvironment.configuration = [
                stages: [
                        myStage: [
                            credentialsId: "testCred"
                        ]
                ]
        ]

        stepRule.step.npmExecuteEndToEndTests(
                script: nullScript,
                stageName: "myStage"
        )

        assertFalse(executedInParallel)
        assert npmExecuteScriptsRule.hasParameter('script', nullScript)
        assert npmExecuteScriptsRule.hasParameter('parameters', [dockerOptions: ['--shm-size 512MB']])
        assert npmExecuteScriptsRule.hasParameter('virtualFrameBuffer', true)
        assert npmExecuteScriptsRule.hasParameter('runScripts', ["ci-e2e"])
        assert binding.hasVariable('e2e_username')
        assert binding.hasVariable('e2e_password')
        assertFalse binding.hasVariable('wdi5_username')
        assertFalse binding.hasVariable('wdi5_password')

    }

    @Test
    void chooseScript() {

        nullScript.commonPipelineEnvironment.configuration = [
                stages: [
                        myStage: [
                                runScript: "wdio"
                        ]
                ]
        ]

        stepRule.step.npmExecuteEndToEndTests(
                script: nullScript,
                stageName: "myStage"
        )

        assertFalse(executedInParallel)
        assert npmExecuteScriptsRule.hasParameter('script', nullScript)
        assert npmExecuteScriptsRule.hasParameter('parameters', [dockerOptions: ['--shm-size 512MB']])
        assert npmExecuteScriptsRule.hasParameter('virtualFrameBuffer', true)
        assert npmExecuteScriptsRule.hasParameter('runScripts', ["wdio"])
        assert npmExecuteScriptsRule.hasParameter('scriptOptions', [])
    }

    @Test
    void oneAppUrlWithCredentials() {
        def appUrl = [url: "http://my-url.com", credentialId: 'testCred']

        nullScript.commonPipelineEnvironment.configuration = [stages: [myStage:[
            appUrls: [appUrl]
        ]]]

        stepRule.step.npmExecuteEndToEndTests(
            script: nullScript,
            stageName: "myStage"
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
            stageName: "myStage"
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
            stageName: "myStage"
        )

        assertTrue(executedInParallel)
        assertTrue(executedOnNode)
        assertFalse(executedOnKubernetes)
    }

    @Test
    void parallelE2eTestOnKubernetes_setWith_POD_NAME_EnvVariable() {
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
            stageName: "myStage"
        )

        assertTrue(executedInParallel)
        assertFalse(executedOnNode)
        assertTrue(executedOnKubernetes)
    }

    @Test
    void parallelE2eTestOnKubernetes_setWith_ON_K8S_EnvVariable() {
        def appUrl = [url: "http://my-url.com", credentialId: 'testCred']
        binding.variables.env.ON_K8S = "true"

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
