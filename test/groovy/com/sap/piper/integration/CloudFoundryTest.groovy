package com.sap.piper.integration

import hudson.AbortException
import org.junit.Rule
import org.junit.Test
import org.junit.Ignore
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import groovy.json.JsonSlurper
import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat
import util.JenkinsCredentialsRule

class CloudFoundryTest extends BasePiperTest {
    public ExpectedException exception = ExpectedException.none()
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsCredentialsRule jenkinsCredentialsRule = new JenkinsCredentialsRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(exception)
        .around(new JenkinsErrorRule(this))
        .around(new JenkinsReadJsonRule(this))
        .around(shellRule)
        .around(jenkinsCredentialsRule)

    @Test
    void getAuthEndPoint_test() {
        helper.registerAllowedMethod('httpRequest', [Map.class] , {
            return [content: '{ "authorization_endpoint": "myAuthEndPoint" }', status: 200]
        })

        String apiEndPoint = 'http://dummy.sap.com'
        boolean verbose = true

        def cf = new CloudFoundry(nullScript)
        def endPoint = cf.getAuthEndPoint(apiEndPoint, verbose)

        assertThat(endPoint, is('myAuthEndPoint'))
    }

    @Test
    void getBearerToken_test(){
        String authorizationEndpoint = 'http://dummy.sap.com' 
        String credentialsId = 'credentialsId' 
        boolean verbose = true

        jenkinsCredentialsRule.withCredentials(credentialsId, 'myuser', 'topsecret')

        helper.registerAllowedMethod('httpRequest', [Map.class] , {
            return [content: '{ "access_token": "myAccessToken" }', status: 200]
        })

        def cf = new CloudFoundry(nullScript)
        def token = cf.getBearerToken(authorizationEndpoint, credentialsId, verbose)

        assertThat(token.toString(), is('Bearer myAccessToken'))
    }

    @Test
    void getAppRefUrl_test(){
        String apiEndpoint =  'http://dummy.sap.com'
        String org = 'myOrg'
        String space = 'mySpace'
        String bearerToken = 'myAccessToken'
        String appName = 'myAppName'
        boolean verbose = true

        helper.registerAllowedMethod('httpRequest', [Map.class] , {
            return [content: '{ "resources":[{"guid":"myGuid", "links":{"self":{"href":"myAppUrl"}}}] }', status: 200]
        })

        def cf = new CloudFoundry(nullScript)
        def appUrl = cf.getAppRefUrl(apiEndpoint, org, space, bearerToken, appName, verbose)

        assertThat(appUrl.toString(), is('myAppUrl'))
    }

    @Test
    void getOrgGuid_test(){
        String apiEndpoint = 'http://dummy.sap.com'
        String org = 'myOrg'
        String bearerToken = 'myToken'
        boolean verbose = true

        helper.registerAllowedMethod('httpRequest', [Map.class] , {
            return [content: '{ "resources":[{"guid":"myOrgGuid"}] }', status: 200]
        })

        def cf = new CloudFoundry(nullScript)
        def orgGuid = cf.getOrgGuid(apiEndpoint, org, bearerToken, verbose)

        assertThat(orgGuid.toString(), is('myOrgGuid'))
    }

    @Test
    void getSpaceGuid_test(){
        String apiEndpoint ='http://dummy.sap.com'
        String orgGuid = 'myOrgGuid'
        String space = 'mySpace'
        String bearerToken = 'myToken'
        boolean verbose = true

        helper.registerAllowedMethod('httpRequest', [Map.class] , {
            return [content: '{ "resources":[{"guid":"mySpaceGuid"}] }', status: 200]
        })

        def cf = new CloudFoundry(nullScript)
        def spaceGuid = cf.getSpaceGuid(apiEndpoint, orgGuid, space, bearerToken, verbose)

        assertThat(spaceGuid.toString(), is('mySpaceGuid'))
    }

    @Test
    void getAppEnvironment_test(){
        String apiEndpoint = 'http://dummy.sap.com'
        String org = 'myOrg'
        String space = 'mySpace'
        String credentialsId = 'credentialsId'
        String appName = 'myAppName'
        boolean verbose = true

        jenkinsCredentialsRule.withCredentials(credentialsId, 'myuser', 'topsecret')

        helper.registerAllowedMethod('httpRequest', [Map.class] , {
            return [content: '{ "authorization_endpoint": "myAuthEndPoint", "access_token": "myAccessToken", "resources":[{"guid":"myGuid", "links":{"self":{"href":"myAppUrl"}}}] }', status: 200]
        })

        def cf = new CloudFoundry(nullScript)
        def appEnv = cf.getAppEnvironment(apiEndpoint, org, space, credentialsId, appName, verbose)

        assertThat(appEnv.toString(), is('{authorization_endpoint=myAuthEndPoint, access_token=myAccessToken, resources=[{guid=myGuid, links={self={href=myAppUrl}}}]}'))
    }

    @Test
    void getXsuaaCredentials_test(){
        String apiEndpoint = 'http://dummy.sap.com'
        String org = 'myOrg'
        String space = 'mySpace'
        String credentialsId = 'credentialsId'
        String appName = 'myAppName'
        boolean verbose = true

        jenkinsCredentialsRule.withCredentials(credentialsId, 'myuser', 'topsecret')

        helper.registerAllowedMethod('httpRequest', [Map.class] , {
            return [content: '{ "system_env_json":{"VCAP_SERVICES":{"xsuaa":[{"credentials":"myCredentials"}]}}, "authorization_endpoint": "myAuthEndPoint", "access_token": "myAccessToken", "resources":[{"guid":"myGuid", "links":{"self":{"href":"myAppUrl"}}}] }', status: 200]
        })

        def cf = new CloudFoundry(nullScript)
        def xsuaaCred = cf.getXsuaaCredentials(apiEndpoint, org, space, credentialsId, appName, verbose)

        assertThat(xsuaaCred.toString(), is('myCredentials'))
    }
}
