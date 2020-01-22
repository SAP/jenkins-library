import java.util.Map
import static org.hamcrest.Matchers.hasItem
import static org.junit.Assert.assertThat

import org.hamcrest.Matchers
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.equalTo
import static org.hamcrest.Matchers.hasEntry
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsStepRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsDockerExecuteRule
import com.sap.piper.JenkinsUtils
import util.Rules

import hudson.AbortException

public class CloudFoundryCreateServiceKeyTest extends BasePiperTest {

    private ExpectedException thrown = new ExpectedException()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellRule = new JenkinsShellCallRule(this)
    private JenkinsDockerExecuteRule dockerExecuteRule = new JenkinsDockerExecuteRule(this)
    private JenkinsCredentialsRule credentialsRule = new JenkinsCredentialsRule(this).withCredentials('test_credentialsId', 'user', 'password')

    class JenkinsUtilsMock extends JenkinsUtils {
        def isJobStartedByUser() {
            return true
        }
    }

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(stepRule)
        .around(loggingRule)
        .around(credentialsRule)
        .around(dockerExecuteRule)
        .around(shellRule)

    @Before
    public void setup() {
    }

    @Test
    public void success() {
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*cf create-service-key.*/, 0 )

        stepRule.step.cloudFoundryCreateServiceKey(
            script: nullScript,
            cloudFoundry: [
                apiEndpoint: 'api.example.com',
                credentialsId: 'test_credentialsId',
                org: 'testOrg',
                space: 'testSpace',
                serviceInstance : 'myInstance',
                serviceKey : 'myServiceKey',
                serviceKeyConfig : '{ "key" : "value" }'
                ]
        )
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))
        assertThat(shellRule.shell, hasItem(containsString("#!/bin/bash set +x set -e export HOME=/home/piper cf login -u 'user' -p 'password' -a api.example.com -o 'testOrg' -s 'testSpace'; cf create-service-key 'myInstance' 'myServiceKey' -c '{ \"key\" : \"value\" }'")))
    }

    @Test
    public void successFlatCloudFoundryParameters() {
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*cf create-service-key.*/, 0 )

        stepRule.step.cloudFoundryCreateServiceKey(
            script: nullScript,
            cfApiEndpoint: 'api.example.com',
            cfCredentialsId: 'test_credentialsId',
            cfOrg: 'testOrg',
            cfSpace: 'testSpace',
            cfServiceInstance : 'myInstance',
            cfServiceKey : 'myServiceKey',
            cfServiceKeyConfig : '{ "key" : "value" }'
        )
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))
        assertThat(shellRule.shell, hasItem(containsString("#!/bin/bash set +x set -e export HOME=/home/piper cf login -u 'user' -p 'password' -a api.example.com -o 'testOrg' -s 'testSpace'; cf create-service-key 'myInstance' 'myServiceKey' -c '{ \"key\" : \"value\" }'")))
    }

    @Test
    public void noServiceKeyConfig() {
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*cf create-service-key.*/, 0 )

        stepRule.step.cloudFoundryCreateServiceKey(
            script: nullScript,
            cloudFoundry: [
                apiEndpoint: 'api.example.com',
                credentialsId: 'test_credentialsId',
                org: 'testOrg',
                space: 'testSpace',
                serviceInstance : 'myInstance',
                serviceKey : 'myServiceKey'
                ]
        )
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))
        assertThat(shellRule.shell, hasItem(containsString("#!/bin/bash set +x set -e export HOME=/home/piper cf login -u 'user' -p 'password' -a api.example.com -o 'testOrg' -s 'testSpace'; cf create-service-key 'myInstance' 'myServiceKey'")))
    }

    public void fail() {
        shellRule.setReturnValue(JenkinsShellCallRule.Type.REGEX, /.*cf create-service-key.*/, 1 )

        stepRule.step.cloudFoundryCreateServiceKey(
            script: nullScript,
            cloudFoundry: [
                apiEndpoint: 'api.example.com',
                credentialsId: 'test_credentialsId',
                org: 'testOrg',
                space: 'testSpace',
                serviceInstance : 'myInstance',
                serviceKey : 'myServiceKey',
                serviceKeyConfig : '{ "key" : "value" }'
                ]
        )
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerImage', 'ppiper/cf-cli'))
        assertThat(dockerExecuteRule.dockerParams, hasEntry('dockerWorkspace', '/home/piper'))
        assertThat(shellRule.shell, hasItem(containsString("[cloudFoundryCreateServiceKey] ERROR: The execution of the create-service-key failed, see the logs above for more details.")))
    }
}
