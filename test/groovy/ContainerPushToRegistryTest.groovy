#!groovy
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsCredentialsRule
import util.JenkinsDockerExecuteRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.Rules

import com.sap.piper.Utils

import static org.hamcrest.CoreMatchers.containsString
import static org.hamcrest.CoreMatchers.hasItem
import static org.hamcrest.CoreMatchers.is
import static org.hamcrest.CoreMatchers.not
import static org.junit.Assert.assertThat

class ContainerPushToRegistryTest extends BasePiperTest {
    private ExpectedException exception = ExpectedException.none()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsDockerExecuteRule dockerRule = new JenkinsDockerExecuteRule(this)
    private JenkinsShellCallRule shellCallRule = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(exception)
        .around(shellCallRule)
        .around(dockerRule)
        .around(new JenkinsCredentialsRule(this)
            .withCredentials('testCredentialsId', 'registryUser', '********')
        )
        .around(stepRule)

    def dockerMockArgs = [:]
    class DockerMock {
        DockerMock(name){
            dockerMockArgs.name = name
        }
        def withRegistry(paramRegistry, paramCredentials, paramClosure){
            dockerMockArgs.paramRegistry = paramRegistry
            dockerMockArgs.paramCredentials = paramCredentials
            return paramClosure()
        }
        def withRegistry(paramRegistry, paramClosure){
            dockerMockArgs.paramRegistryAnonymous = paramRegistry.toString()
            return paramClosure()
        }

        def image(name) {
            dockerMockArgs.name = name
            return new ContainerImageMock()
        }
    }

    def dockerMockPushes = []
    def dockerMockPull = false
    class ContainerImageMock {
        ContainerImageMock(){}
        def push(tag){
            dockerMockPushes.add(tag)
        }
        def push(){
            push('default')
        }

        def pull(){
            dockerMockPull = true
        }
    }

    @Before
    void init() {
        binding.setVariable('docker', new DockerMock('test'))
        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
    }

    @Test
    void testNoImageProvided() {
        exception.expectMessage(containsString('Please provide a dockerImage'))
        stepRule.step.containerPushToRegistry(
            script: nullScript,
            dockerRegistryUrl: 'https://testRegistry',
            dockerCredentialsId: 'testCredentialsId',
        )
    }

    @Test
    void testDefault() {
        stepRule.step.containerPushToRegistry(
            script: nullScript,
            dockerRegistryUrl: 'https://testRegistry',
            dockerCredentialsId: 'testCredentialsId',
            dockerImage: 'testImage:tag',
        )

        assertThat(dockerMockArgs.paramRegistry, is('https://testRegistry'))
        assertThat(dockerMockArgs.paramCredentials, is('testCredentialsId'))
        assertThat(dockerMockArgs.name, is('testImage:tag'))
        assertThat(dockerMockPushes, hasItem('default'))
        assertThat(dockerMockPushes, not(hasItem('latest')))
    }

    @Test
    void testBuildImagePushLatest() {
        def dockerBuildImage = new ContainerImageMock()
        stepRule.step.containerPushToRegistry(
            script: nullScript,
            dockerRegistryUrl: 'https://testRegistry',
            dockerCredentialsId: 'testCredentialsId',
            dockerBuildImage: dockerBuildImage,
            tagLatest: true
        )

        assertThat(dockerMockArgs.paramRegistry, is('https://testRegistry'))
        assertThat(dockerMockArgs.paramCredentials, is('testCredentialsId'))
        assertThat(dockerMockArgs.paramRegistryAnonymous, is(null))
        assertThat(dockerMockArgs.name, is('test'))
        assertThat(dockerMockPushes, hasItem('default'))
        assertThat(dockerMockPushes, hasItem('latest'))
    }



    @Test
    void testFromEnv() {
        nullScript.commonPipelineEnvironment.setValue('containerImage', 'path/testImage:tag')
        nullScript.commonPipelineEnvironment.setValue('containerRegistryUrl', 'https://testRegistry:55555')

        stepRule.step.containerPushToRegistry(
            script: nullScript,
            dockerRegistryUrl: 'https://testRegistry',
            dockerCredentialsId: 'testCredentialsId',
        )

        assertThat(dockerMockArgs.paramRegistryAnonymous, is('https://testRegistry:55555'))
        assertThat(dockerMockArgs.name, is('path/testImage:tag'))
        assertThat(shellCallRule.shell, hasItem('docker tag testRegistry:55555/path/testImage:tag path/testImage:tag'))
        assertThat(dockerMockPull, is(true))
    }

    @Test
    void testWithSourceImageAndRegistry() {
        stepRule.step.containerPushToRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerRegistryUrl: 'https://testRegistry',
            sourceImage: 'testSourceName:testSourceTag',
            sourceRegistryUrl: 'http://testSourceRegistry'
        )

        assertThat(dockerMockArgs.paramRegistryAnonymous, is('http://testSourceRegistry'))
        assertThat(dockerMockArgs.name, is('testSourceName:testSourceTag'))
        assertThat(shellCallRule.shell, hasItem('docker tag testSourceRegistry/testSourceName:testSourceTag testSourceName:testSourceTag'))
        assertThat(dockerMockPull, is(true))
    }

    @Test
    void testWithSourceAndTarget() {
        stepRule.step.containerPushToRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerImage: 'testImage:tag',
            dockerRegistryUrl: 'https://testRegistry',
            sourceImage: 'testSourceName:testSourceTag',
            sourceRegistryUrl: 'http://testSourceRegistry'
        )

        assertThat(dockerMockArgs.paramRegistryAnonymous, is('http://testSourceRegistry'))
        assertThat(dockerMockArgs.paramCredentials, null)
        assertThat(dockerMockArgs.name, is('testSourceName:testSourceTag'))
        assertThat(shellCallRule.shell, hasItem('docker tag testSourceRegistry/testSourceName:testSourceTag testImage:tag'))
        assertThat(dockerMockPull, is(true))
    }

    @Test
    void testWithAuthenticatedSourceAndTarget() {
        stepRule.step.containerPushToRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerImage: 'testImage:tag',
            dockerRegistryUrl: 'https://testRegistry',
            sourceCredentialsId: 'testCredentialsId',
            sourceImage: 'testSourceName:testSourceTag',
            sourceRegistryUrl: 'http://testSourceRegistry'
        )

        assertThat(dockerMockArgs.paramRegistryAnonymous, is('http://testSourceRegistry'))
        assertThat(dockerMockArgs.paramCredentials, is('testCredentialsId'))
        assertThat(dockerMockArgs.name, is('testSourceName:testSourceTag'))
        assertThat(shellCallRule.shell, hasItem('docker tag testSourceRegistry/testSourceName:testSourceTag testImage:tag'))
        assertThat(dockerMockPull, is(true))
    }

    @Test
    void testKubernetesMove() {
        binding.setVariable('docker', null)
        shellCallRule.setReturnValue('docker ps -q > /dev/null', 1)

        stepRule.step.containerPushToRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerImage: 'testImage:tag',
            dockerRegistryUrl: 'https://my.registry:55555',
            skopeoImage: 'skopeo:latest',
            sourceImage: 'sourceImage:sourceTag',
            sourceRegistryUrl: 'https://my.source.registry:44444'
        )

        assertThat(shellCallRule.shell, hasItem('skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/testImage:tag'))
        assertThat(dockerRule.dockerParams.dockerImage, is('skopeo:latest'))
    }

    @Test
    void testKubernetesMoveTagLatest() {
        binding.setVariable('docker', null)
        shellCallRule.setReturnValue('docker ps -q > /dev/null', 1)

        stepRule.step.containerPushToRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerImage: 'testImage:tag',
            dockerRegistryUrl: 'https://my.registry:55555',
            sourceImage: 'sourceImage:sourceTag',
            sourceRegistryUrl: 'https://my.source.registry:44444',
            tagLatest: true
        )

        assertThat(shellCallRule.shell, hasItem('skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/testImage:tag'))
        assertThat(shellCallRule.shell, hasItem('skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/testImage:latest'))
    }

    @Test
    void testKubernetesSourceOnly() {
        binding.setVariable('docker', null)
        shellCallRule.setReturnValue('docker ps -q > /dev/null', 1)

        stepRule.step.containerPushToRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerRegistryUrl: 'https://my.registry:55555',
            sourceImage: 'sourceImage:sourceTag',
            sourceRegistryUrl: 'https://my.source.registry:44444'
        )

        assertThat(shellCallRule.shell, hasItem('skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/sourceImage:sourceTag'))
    }

    @Test
    void testKubernetesSourceRegistryFromEnv() {
        binding.setVariable('docker', null)
        shellCallRule.setReturnValue('docker ps -q > /dev/null', 1)

        nullScript.commonPipelineEnvironment.setValue('containerImage', 'sourceImage:sourceTag')
        nullScript.commonPipelineEnvironment.setValue('containerRegistryUrl', 'https://my.source.registry:44444')

        stepRule.step.containerPushToRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerRegistryUrl: 'https://my.registry:55555',
            sourceImage: 'sourceImage:sourceTag',
        )

        assertThat(shellCallRule.shell, hasItem('skopeo copy --src-tls-verify=false --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/sourceImage:sourceTag'))
    }

    @Test
    void testKubernetesPushTar() {
        binding.setVariable('docker', null)
        shellCallRule.setReturnValue('docker ps -q > /dev/null', 1)

        exception.expectMessage('Only moving images')
        stepRule.step.containerPushToRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerArchive: 'myImage.tar',
            dockerImage: 'testImage:tag',
            dockerRegistryUrl: 'https://my.registry:55555',
        )
    }
}
