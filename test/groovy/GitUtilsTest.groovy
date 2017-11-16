import com.sap.piper.GitUtils
import hudson.AbortException
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException

class GitUtilsTest extends PiperTestBase {

    @Rule
    public ExpectedException thrown = new ExpectedException().none()

    private gitUtils = new GitUtils()

    class MyScmParts {
        def url
        def name
    }


    @Before
    void setUp() {
        super.setUp()
        messages.clear()
        GroovySystem.metaClassRegistry.removeMetaClass(gitUtils.class)
        gitUtils.metaClass.echo = { s -> messages.add(s)}
        gitUtils.metaClass.error = { s -> throw new AbortException(s) }
    }

    @Test
    void retrieveGitCoordinatesWithParametersTest() {
        binding.setVariable('params', [GIT_URL: "github.com:user1/my_repo.git", GIT_BRANCH: "master"])
        def gitCoordinates = withPipeline(defaultPipeline()).execute(gitUtils)
        assert messages[0] == "[INFO] Building 'master@github.com:user1/my_repo.git'."
        assert gitCoordinates.url == "github.com:user1/my_repo.git"
        assert gitCoordinates.branch == "master"
    }

    @Test
    void retrieveGitCoordinatesFromScmTest() {
        gitUtils.metaClass.scm = [userRemoteConfigs: [new MyScmParts([url: "github.com:user1/remote_repo.git"])], branches: [new MyScmParts([name: "feature1"])]]
        binding.setVariable('params', [GIT_URL: null, GIT_BRANCH: null])
        def gitCoordinates = withPipeline(defaultPipeline()).execute(gitUtils)
        assert messages[0] == "[INFO] Parameters 'GIT_URL' and 'GIT_BRANCH' not set in Jenkins job configuration. Assuming application to be built is contained in the same repository as this Jenkinsfile."
        assert messages[1] == "[INFO] Building 'feature1@github.com:user1/remote_repo.git'."
        assert gitCoordinates.url == "github.com:user1/remote_repo.git"
        assert gitCoordinates.branch == "feature1"
    }

    @Test
    void retrieveGitCoordinatesWithNoScmPresentTest() {
        gitUtils.metaClass.retrieveScm = { -> throw new AbortException('SCM not found.')}
        thrown.expect(AbortException)
        thrown.expectMessage("No Source Code Management setup present. If you define the Pipeline directly in the Jenkins job configuration you have to set up parameters GIT_URL and GIT_BRANCH of the application to be built as Jenkins job parameters.")
        binding.setVariable('params', [GIT_URL: null, GIT_BRANCH: null])
        withPipeline(defaultPipeline()).execute(gitUtils)
    }

    @Test
    void retrieveGitCoordinatesOnlyWithUrlTest() {
        gitUtils.metaClass.retrieveScm = { -> throw new AbortException('SCM not found.')}
        thrown.expect(AbortException)
        thrown.expectMessage("Parameter 'GIT_BRANCH' not set in Jenkins job configuration. Either set both GIT_URL and GIT_BRANCH of the application to be built as Jenkins job parameters or put this Jenkinsfile into the same repository as the application to be built.")
        binding.setVariable('params', [GIT_URL: "github.com:user1/my_repo.git", GIT_BRANCH: null])
        withPipeline(defaultPipeline()).execute(gitUtils)
    }

    @Test
    void retrieveGitCoordinatesOnlyWithBranchTest() {
        gitUtils.metaClass.retrieveScm = { -> throw new AbortException('SCM not found.')}
        thrown.expect(AbortException)
        thrown.expectMessage("Parameter 'GIT_URL' not set in Jenkins job configuration. Either set both GIT_URL and GIT_BRANCH of the application to be built as Jenkins job parameters or put this Jenkinsfile into the same repository as the application to be built.")
        binding.setVariable('params', [GIT_URL: null, GIT_BRANCH: "master"])
        withPipeline(defaultPipeline()).execute(gitUtils)
    }

    @Test
    void retrieveGitCoordinatesFromCommonPipelineEnvironmentTest() {
        binding.setVariable('params', [GIT_URL: "github.com:user1/my_repo.git", GIT_BRANCH: "master"])
        def gitCoordinates = withPipeline(cpePipeline()).execute(gitUtils)
        assert messages[0] == "[INFO] Building 'master@github.com:user1/my_repo.git'."
        assert gitCoordinates.url == "github.com:user1/my_repo.git"
        assert gitCoordinates.branch == "master"
    }


    private defaultPipeline(){
        return '''
               import com.sap.piper.Utils
               @Library('piper-library-os')

               execute(gitUtils) {
                 node() {
                   def gitCoordinates = gitUtils.retrieveGitCoordinates(this)
                   return gitCoordinates
                 }
               }
               return this
               '''
    }

    private cpePipeline(){
        return '''
               import com.sap.piper.Utils
               @Library('piper-library-os')

               execute(gitUtils) {
                 node() {
                   gitUtils.retrieveGitCoordinates(this)
                   return commonPipelineEnvironment.getGitCoordinates()
                 }
               }
               return this
               '''
    }
}
