import org.junit.After
import org.junit.Before
import org.junit.Test
import org.junit.Rule
import org.junit.rules.RuleChain

import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.endsWith
import static org.hamcrest.Matchers.startsWith
import groovy.json.JsonSlurper

import static org.junit.Assert.assertThat

import util.BasePiperTest
import util.JenkinsStepRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsDockerExecuteRule
import util.Rules
import org.junit.rules.ExpectedException
import util.JenkinsCredentialsRule

import com.sap.piper.Utils

class NewmanExecuteTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)
    private JenkinsCredentialsRule jenkinsCredentialsRule = new JenkinsCredentialsRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(dockerExecuteRule)
        .around(shellRule)
        .around(jenkinsCredentialsRule)
        .around(loggingRule)
        .around(stepRule) // needs to be activated after dockerExecuteRule, otherwise executeDocker is not mocked

    def gitMap

    @Before
    void init() throws Exception {
        helper.registerAllowedMethod('stash', [String.class], null)
        helper.registerAllowedMethod('git', [Map.class], {m ->
            gitMap = m
        })
        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            def files
            if(map.glob == 'notFound.json')
                files = []
            else if(map.glob == '**/*.postman_collection.json')
                files = [
                    new File("testCollectionsFolder/A.postman_collection.json"),
                    new File("testCollectionsFolder/B.postman_collection.json")
                ]
            else
                files = [new File(map.glob)]
            return files.toArray()
        })
        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
    }

    @Test
    void testExecuteNewmanDefault() throws Exception {
        stepRule.step.newmanExecute(
            script: nullScript,
            juStabUtils: utils,
            newmanCollection: 'testCollection',
            newmanEnvironment: 'testEnvironment',
            newmanGlobals: 'testGlobals'
        )
        // asserts
        assertThat(shellRule.shell, hasItem(endsWith('npm install newman newman-reporter-html --global --quiet')))
        assertThat(shellRule.shell, hasItem(endsWith('newman run \'testCollection\' --environment \'testEnvironment\' --globals \'testGlobals\' --reporters junit,html --reporter-junit-export \'target/newman/TEST-testCollection.xml\' --reporter-html-export \'target/newman/TEST-testCollection.html\'')))
        assertThat(dockerExecuteRule.dockerParams.dockerImage, is('node:lts-stretch'))
        assertThat(loggingRule.log, containsString('[newmanExecute] Found files [testCollection]'))
        assertJobStatusSuccess()
    }

    @Test
    void testDockerFromCustomStepConfiguration() {

        def expectedImage = 'image:test'
        def expectedEnvVars = ['env1': 'value1', 'env2': 'value2']
        def expectedOptions = '--opt1=val1 --opt2=val2 --opt3'
        def expectedWorkspace = '/path/to/workspace'
        
        nullScript.commonPipelineEnvironment.configuration = [steps:[newmanExecute:[
            dockerImage: expectedImage, 
            dockerOptions: expectedOptions,
            dockerEnvVars: expectedEnvVars,
            dockerWorkspace: expectedWorkspace
            ]]]

        stepRule.step.newmanExecute(
            script: nullScript,
            juStabUtils: utils
        )
        
        assert expectedImage == dockerExecuteRule.dockerParams.dockerImage
        assert expectedOptions == dockerExecuteRule.dockerParams.dockerOptions
        assert expectedEnvVars.equals(dockerExecuteRule.dockerParams.dockerEnvVars)
        assert expectedWorkspace == dockerExecuteRule.dockerParams.dockerWorkspace
    }
    
    @Test
    void testGlobalInstall() throws Exception {
        stepRule.step.newmanExecute(
            script: nullScript,
            juStabUtils: utils,
            newmanCollection: 'testCollection',
            newmanEnvironment: 'testEnvironment',
            newmanGlobals: 'testGlobals'
        )
        // asserts
        assertThat(shellRule.shell, hasItem(startsWith('NPM_CONFIG_PREFIX=~/.npm-global ')))
        assertThat(shellRule.shell, hasItem(startsWith('PATH=$PATH:~/.npm-global/bin')))
        assertJobStatusSuccess()
    }

    @Test
    void testExecuteNewmanWithNoCollection() throws Exception {
        thrown.expectMessage('[newmanExecute] No collection found with pattern \'notFound.json\'')

        stepRule.step.newmanExecute(
            script: nullScript,
            juStabUtils: utils,
            newmanCollection: 'notFound.json'
        )
        // asserts
        assertJobStatusFailure()
    }

    @Test
    void testExecuteNewmanFailOnError() throws Exception {
        stepRule.step.newmanExecute(
            script: nullScript,
            juStabUtils: utils,
            newmanCollection: 'testCollection',
            newmanEnvironment: 'testEnvironment',
            newmanGlobals: 'testGlobals',
            dockerImage: 'testImage',
            testRepository: 'testRepo',
            failOnError: false
        )
        // asserts
        assertThat(dockerExecuteRule.dockerParams.dockerImage, is('testImage'))
        assertThat(gitMap.url, is('testRepo'))
        assertThat(shellRule.shell, hasItem(endsWith('newman run \'testCollection\' --environment \'testEnvironment\' --globals \'testGlobals\' --reporters junit,html --reporter-junit-export \'target/newman/TEST-testCollection.xml\' --reporter-html-export \'target/newman/TEST-testCollection.html\' --suppress-exit-code')))
        assertJobStatusSuccess()
    }

    @Test
    void testExecuteNewmanWithFolder() throws Exception {
        stepRule.step.newmanExecute(
            script: nullScript,
            juStabUtils: utils,
            newmanRunCommand: 'run ${config.newmanCollection} --iteration-data testDataFile --reporters junit,html --reporter-junit-export target/newman/TEST-${config.newmanCollection.toString().replace(File.separatorChar,(char)\'_\').tokenize(\'.\').first()}.xml --reporter-html-export target/newman/TEST-${config.newmanCollection.toString().replace(File.separatorChar,(char)\'_\').tokenize(\'.\').first()}.html'
        )
        // asserts
        assertThat(shellRule.shell, hasItem(endsWith('newman run testCollectionsFolder'+File.separatorChar+'A.postman_collection.json --iteration-data testDataFile --reporters junit,html --reporter-junit-export target/newman/TEST-testCollectionsFolder_A.xml --reporter-html-export target/newman/TEST-testCollectionsFolder_A.html')))
        assertThat(shellRule.shell, hasItem(endsWith('newman run testCollectionsFolder'+File.separatorChar+'B.postman_collection.json --iteration-data testDataFile --reporters junit,html --reporter-junit-export target/newman/TEST-testCollectionsFolder_B.xml --reporter-html-export target/newman/TEST-testCollectionsFolder_B.html')))
        assertJobStatusSuccess()
    }

    @Test
    void testExecuteNewmanCfAppsWithSecrets() throws Exception {
        def jsonResponse = '{ "system_env_json":{"VCAP_SERVICES":{"xsuaa":[{"credentials":{"clientid":"myclientid", "clientsecret":"myclientsecret"}}]}}, "authorization_endpoint": "myAuthEndPoint", "access_token": "myAccessToken", "resources":[{"guid":"myGuid", "links":{"self":{"href":"myAppUrl"}}}] }'
        jenkinsCredentialsRule.withCredentials('credentialsId', 'myuser', 'topsecret')
        helper.registerAllowedMethod('httpRequest', [Map.class] , {
            return [content: jsonResponse, status: 200]
        })
        helper.registerAllowedMethod('readJSON', [Map.class] , {
            return new JsonSlurper().parseText(jsonResponse)
        })

        stepRule.step.newmanExecute(
            script: nullScript,
            juStabUtils: utils,
            newmanCollection: 'testCollection',
            newmanEnvironment: 'testEnvironment',
            newmanGlobals: 'testGlobals',
            dockerImage: 'testImage',
            testRepository: 'testRepo',
            failOnError: false,
            cloudFoundry: ["apiEndpoint": "http://fake.endpoint.com", "org":"myOrg", "space": "mySpace", "credentialsId": "credentialsId"],
            cfAppsWithSecrets: ['app1', 'app2']
        )
        // asserts
        assertThat(shellRule.shell, hasItem(endsWith('--env-var app1_clientid=myclientid --env-var app1_clientsecret=myclientsecret --env-var app2_clientid=myclientid --env-var app2_clientsecret=myclientsecret')))
        assertJobStatusSuccess()
    }
}
