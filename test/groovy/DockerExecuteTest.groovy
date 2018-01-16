
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import util.JenkinsConfigRule
import util.JenkinsLoggingRule
import util.JenkinsSetupRule

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertTrue
import static org.junit.Assert.assertFalse

class DockerExecuteTest extends PiperTestBase {
    private DockerMock docker

    @Rule
    public JenkinsSetupRule jsr = new JenkinsSetupRule(this)

    @Rule
    public JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
	
	@Rule
	public JenkinsConfigRule jcr = new JenkinsConfigRule(this)

    int whichDockerReturnValue = 0

    @Before
    void init() {

        docker = new DockerMock()
        binding.setVariable('docker', docker)
        binding.setVariable('Jenkins', [instance: [pluginManager: [plugins: [new PluginMock()]]]])

        helper.registerAllowedMethod('sh', [Map.class], {return whichDockerReturnValue})
    }

    @Test
    void testExecuteInsideDocker() throws Exception {
        def script = loadScript("test/resources/pipelines/dockerExecuteTest/executeInsideDocker.groovy")
        script.execute()
        assertEquals('maven:3.5-jdk-8-alpine', docker.getImageName())
        assertTrue(docker.isImagePulled())
        assertEquals(' --env http_proxy --env https_proxy --env no_proxy --env HTTP_PROXY --env HTTPS_PROXY --env NO_PROXY', docker.getParameters())
        assert jlr.log.contains('Inside Docker')
    }

    @Test
    void testExecuteInsideDockerWithParameters() throws Exception {
        def script = loadScript("test/resources/pipelines/dockerExecuteTest/executeInsideDockerWithParameters.groovy")

        script.execute()
        assertTrue(docker.getParameters().contains(' --env https_proxy '))
        assertTrue(docker.getParameters().contains(' --env http_proxy=http://proxy:8000'))
        assertTrue(docker.getParameters().contains(' -it'))
        assertTrue(docker.getParameters().contains(' --volume my_vol:/my_vol'))
    }

	@Test
	void testDockerNotInstalledResultsInLocalExecution() throws Exception {

        whichDockerReturnValue = 1
        def script = loadScript("test/resources/pipelines/dockerExecuteTest/executeInsideDockerWithParameters.groovy")

        script.execute()
        assert jlr.log.contains('No docker environment found')
        assert jlr.log.contains('Running on local environment')
        assertFalse(docker.isImagePulled())
    }

    private class DockerMock {
        private String imageName
        private boolean imagePulled = false
        private String parameters

        DockerMock image(String imageName) {
            this.imageName = imageName
            return this
        }

        void pull() {
            imagePulled = true
        }

        void inside(String parameters, body) {
            this.parameters = parameters
            body()
        }

        String getImageName() {
            return imageName
        }

        boolean isImagePulled() {
            return imagePulled
        }

        String getParameters() {
            return parameters
        }
    }

    private class PluginMock {
        def getShortName() {
            return 'docker-workflow'
        }
        boolean isActive() {
            return true
        }
    }

}
