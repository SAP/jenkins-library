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

    def dockerMock = new DockerMock('test')

    class DockerMock {

        LinkedList<DockerRegistry> registries
        String image

        DockerMock(name){
            this.image = name
            this.registries = [] as LinkedList
        }
        def withRegistry(paramRegistry, paramCredentials, paramClosure){
            this.registries.add(new DockerRegistry(paramRegistry, paramCredentials, paramCredentials? false : true))
            return paramClosure()
        }
        def withRegistry(paramRegistry, paramClosure){
            return withRegistry(paramRegistry, null, paramClosure)
        }

        def image(name) {
            this.image = name
            return new ContainerImageMock()
        }

        protected DockerRegistry getSourceRegistry() {
            if (registries.size() == 1) {
                return null;
            }
            return registries.first
        }
        protected DockerRegistry getTargetRegistry() {
            if (registries.size() == 1) {
                return registries.first
            }
            return registries.last
        }
    }

    class DockerRegistry {
        final boolean isAnonymous
        final String credentials
        final String url

        DockerRegistry(url, credentials, isAnonymous) {
            this.url = url
            this.credentials = credentials
            this.isAnonymous = isAnonymous
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
        binding.setVariable('docker', dockerMock)
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

        assertThat(dockerMock.targetRegistry.url, is('https://testRegistry'))
        assertThat(dockerMock.targetRegistry.credentials, is('testCredentialsId'))
        assertThat(dockerMock.image, is('testImage:tag'))
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

        assertThat(dockerMock.targetRegistry.url, is('https://testRegistry'))
        assertThat(dockerMock.targetRegistry.credentials, is('testCredentialsId'))
        assertThat(dockerMock.sourceRegistry, is (null))
        assertThat(dockerMock.image, is('test'))
        assertThat(dockerMockPushes, hasItem('default'))
        assertThat(dockerMockPushes, hasItem('latest'))
    }

    @Test
    void testBuildImagePushArtifactVersion() throws Exception {
        nullScript.commonPipelineEnvironment.setArtifactVersion('1.0.0')
        def dockerBuildImage = new ContainerImageMock()
        stepRule.step.containerPushToRegistry(
            script: nullScript,
            dockerRegistryUrl: 'https://testRegistry',
            dockerCredentialsId: 'testCredentialsId',
            dockerBuildImage: dockerBuildImage,
            tagArtifactVersion: true
        )

        assertThat(dockerMock.targetRegistry.url, is('https://testRegistry'))
        assertThat(dockerMock.targetRegistry.credentials, is('testCredentialsId'))
        assertThat(dockerMock.sourceRegistry, is (null))
        assertThat(dockerMock.image, is('test'))
        assertThat(dockerMockPushes, hasItem('default'))
        assertThat(dockerMockPushes, not(hasItem('latest')))
        assertThat(dockerMockPushes, hasItem('1.0.0'))
    }

    @Test
    void testBuildImagePushLatestAndArtifactVersion() throws Exception {
        nullScript.commonPipelineEnvironment.setArtifactVersion('1.0.0')
        def dockerBuildImage = new ContainerImageMock()
        stepRule.step.containerPushToRegistry(
            script: nullScript,
            dockerRegistryUrl: 'https://testRegistry',
            dockerCredentialsId: 'testCredentialsId',
            dockerBuildImage: dockerBuildImage,
            tagArtifactVersion: true,
            tagLatest: true
        )

        assertThat(dockerMock.targetRegistry.url, is('https://testRegistry'))
        assertThat(dockerMock.targetRegistry.credentials, is('testCredentialsId'))
        assertThat(dockerMock.sourceRegistry, is (null))
        assertThat(dockerMock.image, is('test'))
        assertThat(dockerMockPushes, hasItem('default'))
        assertThat(dockerMockPushes, hasItem('latest'))
        assertThat(dockerMockPushes, hasItem('1.0.0'))
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

        assertThat(dockerMock.sourceRegistry.url, is('https://testRegistry:55555'))
        assertThat(dockerMock.sourceRegistry.isAnonymous, is(true))
        assertThat(dockerMock.targetRegistry.url, is('https://testRegistry'))
        assertThat(dockerMock.targetRegistry.credentials, is('testCredentialsId'))
        assertThat(dockerMock.targetRegistry.isAnonymous, is(false))
        assertThat(dockerMock.image, is('path/testImage:tag'))
        assertThat(shellCallRule.shell, hasItem("docker tag 'testRegistry:55555'/'path/testImage:tag' 'path/testImage:tag'"))
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

        assertThat(dockerMock.sourceRegistry.url, is('http://testSourceRegistry'))
        assertThat(dockerMock.image, is('testSourceName:testSourceTag'))
        assertThat(dockerMock.sourceRegistry.isAnonymous, is(true))
        assertThat(shellCallRule.shell, hasItem("docker tag 'testSourceRegistry'/'testSourceName:testSourceTag' 'testSourceName:testSourceTag'"))
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

        assertThat(dockerMock.sourceRegistry.url, is('http://testSourceRegistry'))
        assertThat(dockerMock.sourceRegistry.isAnonymous, is(true))
        assertThat(dockerMock.image, is('testSourceName:testSourceTag'))
        assertThat(shellCallRule.shell, hasItem("docker tag 'testSourceRegistry'/'testSourceName:testSourceTag' 'testImage:tag'"))
        assertThat(dockerMockPull, is(true))
    }

    @Test
    void testWithAuthenticatedSourceAndTarget() {
        stepRule.step.containerPushToRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerImage: 'testImage:tag',
            dockerRegistryUrl: 'https://testRegistry',
            sourceCredentialsId: 'testSourceCredentialsId',
            sourceImage: 'testSourceName:testSourceTag',
            sourceRegistryUrl: 'http://testSourceRegistry'
        )

        assertThat(dockerMock.sourceRegistry.url, is('http://testSourceRegistry'))
        assertThat(dockerMock.sourceRegistry.credentials, is('testSourceCredentialsId'))
        assertThat(dockerMock.targetRegistry.url, is('https://testRegistry'))
        assertThat(dockerMock.targetRegistry.credentials, is('testCredentialsId'))
        assertThat(dockerMock.image, is('testSourceName:testSourceTag'))
        assertThat(shellCallRule.shell, hasItem("docker tag 'testSourceRegistry'/'testSourceName:testSourceTag' 'testImage:tag'"))
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
            sourceRegistryUrl: 'https://my.source.registry:44444',
        )

        assertThat(shellCallRule.shell, hasItem('skopeo copy --multi-arch=all --src-tls-verify=false --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/testImage:tag'))
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
            sourceCredentialsId: 'testCredentialsId',
            tagLatest: true
        )

        assertThat(shellCallRule.shell, hasItem('skopeo copy --multi-arch=all --src-tls-verify=false --src-creds=\'registryUser\':\'********\' --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/testImage:tag'))
        assertThat(shellCallRule.shell, hasItem('skopeo copy --multi-arch=all --src-tls-verify=false --src-creds=\'registryUser\':\'********\' --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/testImage:latest'))
    }

    @Test
    void testKubernetesMoveTagArtifactVersion() {
        nullScript.commonPipelineEnvironment.setArtifactVersion('1.0.0')
        binding.setVariable('docker', null)
        shellCallRule.setReturnValue('docker ps -q > /dev/null', 1)

        stepRule.step.containerPushToRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerImage: 'testImage:tag',
            dockerRegistryUrl: 'https://my.registry:55555',
            sourceImage: 'sourceImage:sourceTag',
            sourceRegistryUrl: 'https://my.source.registry:44444',
            sourceCredentialsId: 'testCredentialsId',
            tagArtifactVersion: true
        )

        assertThat(shellCallRule.shell, hasItem('skopeo copy --multi-arch=all --src-tls-verify=false --src-creds=\'registryUser\':\'********\' --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/testImage:tag'))
        assertThat(shellCallRule.shell, hasItem('skopeo copy --multi-arch=all --src-tls-verify=false --src-creds=\'registryUser\':\'********\' --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/testImage:1.0.0'))
    }

    @Test
    void testKubernetesMoveTagLatestAndArtifactVersion() {
        nullScript.commonPipelineEnvironment.setArtifactVersion('1.0.0')
        binding.setVariable('docker', null)
        shellCallRule.setReturnValue('docker ps -q > /dev/null', 1)

        stepRule.step.containerPushToRegistry(
            script: nullScript,
            dockerCredentialsId: 'testCredentialsId',
            dockerImage: 'testImage:tag',
            dockerRegistryUrl: 'https://my.registry:55555',
            sourceImage: 'sourceImage:sourceTag',
            sourceRegistryUrl: 'https://my.source.registry:44444',
            sourceCredentialsId: 'testCredentialsId',
            tagLatest: true,
            tagArtifactVersion: true
        )

        assertThat(shellCallRule.shell, hasItem('skopeo copy --multi-arch=all --src-tls-verify=false --src-creds=\'registryUser\':\'********\' --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/testImage:tag'))
        assertThat(shellCallRule.shell, hasItem('skopeo copy --multi-arch=all --src-tls-verify=false --src-creds=\'registryUser\':\'********\' --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/testImage:latest'))
        assertThat(shellCallRule.shell, hasItem('skopeo copy --multi-arch=all --src-tls-verify=false --src-creds=\'registryUser\':\'********\' --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/testImage:1.0.0'))
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
            sourceRegistryUrl: 'https://my.source.registry:44444',
        )

        assertThat(shellCallRule.shell, hasItem('skopeo copy --multi-arch=all --src-tls-verify=false --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/sourceImage:sourceTag'))
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
            sourceCredentialsId: 'testCredentialsId'
        )

        assertThat(shellCallRule.shell, hasItem('skopeo copy --multi-arch=all --src-tls-verify=false --src-creds=\'registryUser\':\'********\' --dest-tls-verify=false --dest-creds=\'registryUser\':\'********\' docker://my.source.registry:44444/sourceImage:sourceTag docker://my.registry:55555/sourceImage:sourceTag'))
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
