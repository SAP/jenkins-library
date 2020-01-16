import java.util.Map
import static org.hamcrest.Matchers.hasItem
import static org.junit.Assert.assertThat

import org.hamcrest.Matchers
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.equalTo
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsStepRule
import util.JenkinsReadJsonRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsDockerExecuteRule
import util.JenkinsShellCallRule
import util.Rules

import hudson.AbortException

public class AbapEnvironmentPullGitRepoTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsReadJsonRule readJsonRule = new JenkinsReadJsonRule(this)
    private JenkinsCredentialsRule credentialsRule = new JenkinsCredentialsRule(this).withCredentials('test_credentialsId', 'user', 'password')

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(dockerExecuteRule)
        .around(stepRule)
        .around(loggingRule)
        .around(readJsonRule)
        .around(credentialsRule)
        .around(shellRule)

    @Before
    public void setup() {
        UUID.metaClass.static.randomUUID = { -> 1}
    }

    @After
    public void tearDown() {
        UUID.metaClass = null
    }

    @Test
    public void pullSuccessfulCredentialsId() {
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*x-csrf-token: fetch.*/, null )
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*POST.*/, /{"d" : { "__metadata" : { "uri" : "https:\/\/example.com\/URI" } , "status" : "R", "status_descr" : "RUNNING" }}/)
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*https:\/\/example\.com.*/, /{"d" : { "__metadata" : { "uri" : "https:\/\/example.com\/URI" } , "status" : "S", "status_descr" : "SUCCESS" }}/)

        helper.registerAllowedMethod("readFile", [String.class], {
            /HTTP\/1.1 200 OK
            set-cookie: sap-usercontext=sap-client=100; path=\/
            content-type: application\/json; charset=utf-8
            x-csrf-token: TOKEN/
        })

        loggingRule.expect("[abapEnvironmentPullGitRepo] Info: Using configuration: credentialsId: test_credentialsId and host: example.com")
        loggingRule.expect("[abapEnvironmentPullGitRepo] Pull Status: RUNNING")
        loggingRule.expect("[abapEnvironmentPullGitRepo] Entity URI: https://example.com/URI")
        loggingRule.expect("[abapEnvironmentPullGitRepo] Pull Status: SUCCESS")

        stepRule.step.abapEnvironmentPullGitRepo(script: nullScript, host: 'example.com', repositoryName: 'Z_DEMO_DM', credentialsId: 'test_credentialsId')

        assertThat(shellRule.shell[0], containsString(/#!\/bin\/bash curl -I -X GET https:\/\/example.com\/sap\/opu\/odata\/sap\/MANAGE_GIT_REPOSITORY\/Pull -H 'Authorization: Basic dXNlcjpwYXNzd29yZA==' -H 'Accept: application\/json' -H 'x-csrf-token: fetch' -D headerFileAuth-1.txt/))
        assertThat(shellRule.shell[1], containsString(/#!\/bin\/bash curl -X POST "https:\/\/example.com\/sap\/opu\/odata\/sap\/MANAGE_GIT_REPOSITORY\/Pull" -H 'Authorization: Basic dXNlcjpwYXNzd29yZA==' -H 'Accept: application\/json' -H 'Content-Type: application\/json' -H 'x-csrf-token: TOKEN' --cookie headerFileAuth-1.txt -D headerFilePost-1.txt -d '{ "sc_name": "Z_DEMO_DM" }'/))
        assertThat(shellRule.shell[2], containsString(/#!\/bin\/bash curl -X GET "https:\/\/example.com\/URI" -H 'Authorization: Basic dXNlcjpwYXNzd29yZA==' -H 'Accept: application\/json' -D headerFilePoll-1.txt/))
        assertThat(shellRule.shell[3], containsString(/#!\/bin\/bash rm -f headerFileAuth-1.txt headerFilePost-1.txt headerFilePoll-1.txt/))
    }

    @Test
    public void pullSuccessfulCloudFoundry() {
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*cf service-key.*/, 0 )
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*x-csrf-token: fetch.*/, null )
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*POST.*/, /{"d" : { "__metadata" : { "uri" : "https:\/\/example.com\/URI" } , "status" : "R", "status_descr" : "RUNNING" }}/)
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*https:\/\/example.com.*/, /{"d" : { "__metadata" : { "uri" : "https:\/\/example.com\/URI" } , "status" : "S", "status_descr" : "SUCCESS" }}/)

        helper.registerAllowedMethod("readFile", [String.class], { input ->
            if (input.contains("response")) {
                /Getting key SK_NAME_4 for service instance D09_TEST as P2001217173...\/

                {
                "abap": {
                "password": "password",
                "username": "user"
                },
                "url": "https:\/\/example.com"
                }/
            } else {
                /HTTP\/1.1 200 OK
                set-cookie: sap-usercontext=sap-client=100; path=\/
                content-type: application\/json; charset=utf-8
                x-csrf-token: TOKEN/
            }
        })

        loggingRule.expect("[abapEnvironmentPullGitRepo] Info: Using Cloud Foundry service key testKey for service instance testInstance")
        loggingRule.expect("[abapEnvironmentPullGitRepo] Pull Status: RUNNING")
        loggingRule.expect("[abapEnvironmentPullGitRepo] Entity URI: https://example.com/URI")
        loggingRule.expect("[abapEnvironmentPullGitRepo] Pull Status: SUCCESS")

        stepRule.step.abapEnvironmentPullGitRepo(
            script: nullScript,
            repositoryName: 'Z_DEMO_DM',
            cloudFoundry: [
                apiEndpoint : 'api.cloudfoundry.com',
                org : 'testOrg',
                space : 'testSpace',
                credentialsId : 'test_credentialsId',
                serviceInstance : 'testInstance',
                serviceKey : 'testKey'
            ])

        assertThat(shellRule.shell[0], containsString(/#!\/bin\/bash set +x set -e export HOME=\/home\/piper cf login -u 'user' -p 'password' -a api.cloudfoundry.com -o 'testOrg' -s 'testSpace'; cf service-key 'testInstance' 'testKey' > "response-1.txt/))
        assertThat(shellRule.shell[1], containsString(/cf logout/))
        assertThat(shellRule.shell[2], containsString(/#!\/bin\/bash rm -f response-1.txt/))
        assertThat(shellRule.shell[3], containsString(/#!\/bin\/bash curl -I -X GET https:\/\/example.com\/sap\/opu\/odata\/sap\/MANAGE_GIT_REPOSITORY\/Pull -H 'Authorization: Basic dXNlcjpwYXNzd29yZA==' -H 'Accept: application\/json' -H 'x-csrf-token: fetch' -D headerFileAuth-1.txt/))
        assertThat(shellRule.shell[4], containsString(/#!\/bin\/bash curl -X POST "https:\/\/example.com\/sap\/opu\/odata\/sap\/MANAGE_GIT_REPOSITORY\/Pull" -H 'Authorization: Basic dXNlcjpwYXNzd29yZA==' -H 'Accept: application\/json' -H 'Content-Type: application\/json' -H 'x-csrf-token: TOKEN' --cookie headerFileAuth-1.txt -D headerFilePost-1.txt -d '{ "sc_name": "Z_DEMO_DM" }'/))
        assertThat(shellRule.shell[5], containsString(/#!\/bin\/bash curl -X GET "https:\/\/example.com\/URI" -H 'Authorization: Basic dXNlcjpwYXNzd29yZA==' -H 'Accept: application\/json' -D headerFilePoll-1.txt/))
        assertThat(shellRule.shell[6], containsString(/#!\/bin\/bash rm -f headerFileAuth-1.txt headerFilePost-1.txt headerFilePoll-1.txt/))
    }

    @Test
    public void pullSuccessfulCloudFoundryFlatParameters() {
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*cf service-key.*/, 0 )
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*x-csrf-token: fetch.*/, null )
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*POST.*/, /{"d" : { "__metadata" : { "uri" : "https:\/\/example.com\/URI" } , "status" : "R", "status_descr" : "RUNNING" }}/)
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*https:\/\/example.com.*/, /{"d" : { "__metadata" : { "uri" : "https:\/\/example.com\/URI" } , "status" : "S", "status_descr" : "SUCCESS" }}/)

        helper.registerAllowedMethod("readFile", [String.class], { input ->
            if (input.contains("response")) {
                /Getting key SK_NAME_4 for service instance D09_TEST as P2001217173...\/

                {
                "abap": {
                "password": "password",
                "username": "user"
                },
                "url": "https:\/\/example.com"
                }/
            } else {
                /HTTP\/1.1 200 OK
                set-cookie: sap-usercontext=sap-client=100; path=\/
                content-type: application\/json; charset=utf-8
                x-csrf-token: TOKEN/
            }
        })

        loggingRule.expect("[abapEnvironmentPullGitRepo] Info: Using Cloud Foundry service key testKey for service instance testInstance")
        loggingRule.expect("[abapEnvironmentPullGitRepo] Pull Status: RUNNING")
        loggingRule.expect("[abapEnvironmentPullGitRepo] Entity URI: https://example.com/URI")
        loggingRule.expect("[abapEnvironmentPullGitRepo] Pull Status: SUCCESS")

        stepRule.step.abapEnvironmentPullGitRepo(
            script: nullScript,
            repositoryName: 'Z_DEMO_DM',
            cfApiEndpoint : 'api.cloudfoundry.com',
            cfOrg : 'testOrg',
            cfSpace : 'testSpace',
            cfCredentialsId : 'test_credentialsId',
            cfServiceInstance : 'testInstance',
            cfServiceKey : 'testKey'
            )

        assertThat(shellRule.shell[0], containsString(/#!\/bin\/bash set +x set -e export HOME=\/home\/piper cf login -u 'user' -p 'password' -a api.cloudfoundry.com -o 'testOrg' -s 'testSpace'; cf service-key 'testInstance' 'testKey' > "response-1.txt/))
        assertThat(shellRule.shell[1], containsString(/cf logout/))
        assertThat(shellRule.shell[2], containsString(/#!\/bin\/bash rm -f response-1.txt/))
        assertThat(shellRule.shell[3], containsString(/#!\/bin\/bash curl -I -X GET https:\/\/example.com\/sap\/opu\/odata\/sap\/MANAGE_GIT_REPOSITORY\/Pull -H 'Authorization: Basic dXNlcjpwYXNzd29yZA==' -H 'Accept: application\/json' -H 'x-csrf-token: fetch' -D headerFileAuth-1.txt/))
        assertThat(shellRule.shell[4], containsString(/#!\/bin\/bash curl -X POST "https:\/\/example.com\/sap\/opu\/odata\/sap\/MANAGE_GIT_REPOSITORY\/Pull" -H 'Authorization: Basic dXNlcjpwYXNzd29yZA==' -H 'Accept: application\/json' -H 'Content-Type: application\/json' -H 'x-csrf-token: TOKEN' --cookie headerFileAuth-1.txt -D headerFilePost-1.txt -d '{ "sc_name": "Z_DEMO_DM" }'/))
        assertThat(shellRule.shell[5], containsString(/#!\/bin\/bash curl -X GET "https:\/\/example.com\/URI" -H 'Authorization: Basic dXNlcjpwYXNzd29yZA==' -H 'Accept: application\/json' -D headerFilePoll-1.txt/))
        assertThat(shellRule.shell[6], containsString(/#!\/bin\/bash rm -f headerFileAuth-1.txt headerFilePost-1.txt headerFilePoll-1.txt/))
    }

    @Test
    public void pullFailsWhilePolling() {
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*x-csrf-token: fetch.*/, "TOKEN")
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*POST.*/, /{"d" : { "__metadata" : { "uri" : "https:\/\/example.com\/URI" } , "status" : "R", "status_descr" : "RUNNING" }}/)
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*https:\/\/example\.com.*/, /{"d" : { "__metadata" : { "uri" : "https:\/\/example.com\/URI" } , "status" : "E", "status_descr" : "ERROR" }}/)

        helper.registerAllowedMethod("readFile", [String.class], {
            /HTTP\/1.1 200 OK
            set-cookie: sap-usercontext=sap-client=100; path=\/
            content-type: application\/json; charset=utf-8/
        })

        loggingRule.expect("[abapEnvironmentPullGitRepo] Pull Status: RUNNING")
        loggingRule.expect("[abapEnvironmentPullGitRepo] Pull Status: ERROR")

        thrown.expect(Exception)
        thrown.expectMessage("[abapEnvironmentPullGitRepo] Pull Failed")

        stepRule.step.abapEnvironmentPullGitRepo(script: nullScript, host: 'example.com', repositoryName: 'Z_DEMO_DM', credentialsId: 'test_credentialsId')

    }

    @Test
    public void pullFailsWithPostRequest() {
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*x-csrf-token: fetch.*/, "TOKEN")
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*POST.*/, /{"d" : { "__metadata" : { "uri" : "https:\/\/example.com\/URI" } , "status" : "E", "status_descr" : "ERROR" }}/)
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*https:\/\/example\.com.*/, /{"d" : { "__metadata" : { "uri" : "https:\/\/example.com\/URI" } , "status" : "E", "status_descr" : "ERROR" }}/)

        helper.registerAllowedMethod("readFile", [String.class], {
            /HTTP\/1.1 200 OK
            set-cookie: sap-usercontext=sap-client=100; path=\/
            content-type: application\/json; charset=utf-8/
        })

        loggingRule.expect("[abapEnvironmentPullGitRepo] Pull Status: ERROR")

        thrown.expect(Exception)
        thrown.expectMessage("[abapEnvironmentPullGitRepo] Pull Failed")

        stepRule.step.abapEnvironmentPullGitRepo(script: nullScript, host: 'example.com', repositoryName: 'Z_DEMO_DM', credentialsId: 'test_credentialsId')

    }

    @Test
    public void pullWithErrorResponse() {
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*x-csrf-token: fetch.*/, "TOKEN")
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*POST.*/, /{"error" : { "message" : { "lang" : "en", "value": "text" } }}/)

        helper.registerAllowedMethod("readFile", [String.class], {
            /HTTP\/1.1 200 OK
            set-cookie: sap-usercontext=sap-client=100; path=\/
            content-type: application\/json; charset=utf-8/
        })

        thrown.expect(Exception)
        thrown.expectMessage("[abapEnvironmentPullGitRepo] Error: text")

        stepRule.step.abapEnvironmentPullGitRepo(script: nullScript, host: 'example.com', repositoryName: 'Z_DEMO_DM', credentialsId: 'test_credentialsId')

    }

    @Test
    public void connectionFails() {
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*x-csrf-token: fetch.*/, null)

        helper.registerAllowedMethod("readFile", [String.class], {
            /HTTP\/1.1 401 Unauthorized
            set-cookie: sap-usercontext=sap-client=100; path=\/
            content-type: application\/json; charset=utf-8/
        })

        thrown.expect(Exception)
        thrown.expectMessage("[abapEnvironmentPullGitRepo] Error: 401 Unauthorized")

        stepRule.step.abapEnvironmentPullGitRepo(script: nullScript, host: 'example.com', repositoryName: 'Z_DEMO_DM', credentialsId: 'test_credentialsId')

    }

    @Test
    public void checkRepositoryProvided() {
       thrown.expect(IllegalArgumentException)
       thrown.expectMessage("ERROR - NO VALUE AVAILABLE FOR repositoryName")
       stepRule.step.abapEnvironmentPullGitRepo(script: nullScript, host: 'example.com', credentialsId: 'test_credentialsId')
    }

    @Test
    public void testHttpHeader() {

        String header = /HTTP\/1.1 401 Unauthorized
            set-cookie: sap-usercontext=sap-client=100; path=\/
            content-type: text\/html; charset=utf-8
            content-length: 9321
            sap-system: Y11
            x-csrf-token: TOKEN
            www-authenticate: Basic realm="SAP NetWeaver Application Server [Y11\/100][alias]"
            sap-server: true
            sap-perf-fesrec: 72927.000000/

        HttpHeaderProperties httpHeader = new HttpHeaderProperties(header)
        assertThat(httpHeader.statusCode, equalTo(401))
        assertThat(httpHeader.statusMessage, containsString("Unauthorized"))
        assertThat(httpHeader.xCsrfToken, containsString("TOKEN"))
    }
}
