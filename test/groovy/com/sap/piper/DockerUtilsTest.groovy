package com.sap.piper

import hudson.AbortException
import org.junit.Before
import org.junit.Ignore
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsShellCallRule
import util.Rules

import static org.hamcrest.CoreMatchers.hasItem
import static org.hamcrest.CoreMatchers.is
import static org.junit.Assert.assertThat

class DockerUtilsTest extends BasePiperTest {

    public ExpectedException exception = ExpectedException.none()
    public JenkinsShellCallRule shellCallRule = new JenkinsShellCallRule(this)

    def dockerMockArgs = [:]
    class DockerMock {
        def withRegistry(paramRegistry, paramClosure){
            dockerMockArgs.paramRegistryAnonymous = paramRegistry.toString()
            return paramClosure()
        }
    }

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(shellCallRule)
        .around(exception)
        .around(new JenkinsCredentialsRule(this)
            .withCredentials('testCredentialsId', 'registryUser', '********')
        )
    @Before
    void init() {
        nullScript.binding.setVariable('docker', new DockerMock())
    }

    @Test
    void testWithDockerDaemon() {
        DockerUtils dockerUtils = new DockerUtils(nullScript)
        assertThat(dockerUtils.withDockerDaemon(), is(true))
    }

    @Test
    void testWithoutDockerDaemon() {
        shellCallRule.setReturnValue('docker ps -q > /dev/null', 1)
        DockerUtils dockerUtils = new DockerUtils(nullScript)
        assertThat(dockerUtils.withDockerDaemon(), is(false))
    }

    @Test
    void testOnKubernetes() {
        nullScript.env.ON_K8S = 'true'
        DockerUtils dockerUtils = new DockerUtils(nullScript)
        assertThat(dockerUtils.onKubernetes(), is(true))
    }

    @Test
    void testMoveImageKubernetes() {
        shellCallRule.setReturnValue('docker ps -q > /dev/null', 1)
        DockerUtils dockerUtils = new DockerUtils(nullScript)
        dockerUtils.moveImage(
            [
                registryUrl: 'https://my.source.registry:44444',
                image: 'sourceImage:sourceTag',
                credentialsId: 'testCredentialsId'
            ],
            [
                registryUrl: 'https://my.registry:55555',
                image: 'testImage:tag',
                credentialsId: 'testCredentialsId'
            ]
        )

        assertThat(shellCallRule.shell, hasItem('skopeo copy --multi-arch=all --src-tls-verify=false --src-creds=\'registryUser\':\'********\' --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/testImage:tag'))
    }

    @Test
    void testGetRegistryFromUrl() {
        DockerUtils dockerUtils = new DockerUtils(nullScript)
        assertThat(dockerUtils.getRegistryFromUrl('https://my.registry.com:55555'), is('my.registry.com:55555'))
        assertThat(dockerUtils.getRegistryFromUrl('http://my.registry.com:55555'), is('my.registry.com:55555'))
        assertThat(dockerUtils.getRegistryFromUrl('https://my.registry.com'), is('my.registry.com'))
    }

    @Test
    void testGetProtocolFromUrl() {
        DockerUtils dockerUtils = new DockerUtils(nullScript)
        assertThat(dockerUtils.getProtocolFromUrl('https://my.registry.com:55555'), is('https'))
        assertThat(dockerUtils.getProtocolFromUrl('http://my.registry.com:55555'), is('http'))
    }
}
