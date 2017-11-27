import junit.framework.TestCase
import org.junit.Before
import org.junit.Test

class DockerExecuteTest extends AbstractPiperUnitTest {
    private DockerMock docker

    String echos

    @Before
    void setUp() {
        super.setUp()

        docker = new DockerMock()
        binding.setVariable('docker', docker)

        echos = ''
        helper.registerAllowedMethod("echo", [String.class], { String s -> echos += " $s" })
    }

    @Test
    void testExecuteInsideDocker() throws Exception {
        def script = loadScript("test/resources/pipelines/dockerExecuteTest/executeInsideDocker.groovy",)
            script.execute()
            TestCase.assertEquals('maven:3.5-jdk-8-alpine', docker.getImageName())
            TestCase.assertTrue(docker.isImagePulled())
            TestCase.assertEquals(' --env http_proxy --env https_proxy --env no_proxy --env HTTP_PROXY --env HTTPS_PROXY --env NO_PROXY', docker.getParameters())
            TestCase.assertTrue(echos.contains('Inside Docker'))
    }

    @Test
    void testExecuteInsideDockerWithParameters() throws Exception {
        def script = loadScript("test/resources/pipelines/dockerExecuteTest/executeInsideDockerWithParameters.groovy",)

        script.execute()
        TestCase.assertTrue(docker.getParameters().contains(' --env https_proxy '))
        TestCase.assertTrue(docker.getParameters().contains(' --env http_proxy=http://proxy:8000'))
        TestCase.assertTrue(docker.getParameters().contains(' -it'))
        TestCase.assertTrue(docker.getParameters().contains(' --volume my_vol:/my_vol'))
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
}
