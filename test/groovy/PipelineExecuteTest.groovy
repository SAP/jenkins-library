import hudson.AbortException
import util.JenkinsConfigRule
import util.JenkinsSetupRule

import org.junit.rules.TemporaryFolder
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException

class PipelineExecuteTest extends PiperTestBase {

    @Rule
    public ExpectedException thrown = new ExpectedException().none()

    @Rule
    public JenkinsSetupRule jsr = new JenkinsSetupRule(this)
	
	@Rule
	public JenkinsConfigRule jcr = new JenkinsConfigRule(this)

    def pipelinePath
    def checkoutParameters = [:]
    def load

    @Before
    void init() {

        pipelinePath = null
        checkoutParameters.clear()
        load = null

        helper.registerAllowedMethod('deleteDir', [], null)
        helper.registerAllowedMethod('checkout', [Map], { m ->
            checkoutParameters.branch = m.branches[0].name
            checkoutParameters.repoUrl = m.userRemoteConfigs[0].url
            checkoutParameters.credentialsId = m.userRemoteConfigs[0].credentialsId
            checkoutParameters.path = m.extensions[0].sparseCheckoutPaths[0].path
        })
        helper.registerAllowedMethod('load', [String], { s -> load = s })

    }


    @Test
    void straightForwardTest() {

        withPipeline(defaultPipeline()).execute()
        assert load == "Jenkinsfile"
        assert checkoutParameters.branch == 'master'
        assert checkoutParameters.repoUrl == "https://test.com/myRepo.git"
        assert checkoutParameters.credentialsId == ''
        assert checkoutParameters.path == 'Jenkinsfile'

    }

    @Test
    void parameterizeTest() {

        withPipeline(parameterizePipeline()).execute()
        assert load == "path/to/Jenkinsfile"
        assert checkoutParameters.branch == 'feature'
        assert checkoutParameters.repoUrl == "https://test.com/anotherRepo.git"
        assert checkoutParameters.credentialsId == 'abcd1234'
        assert checkoutParameters.path == 'path/to/Jenkinsfile'

    }

    @Test
    void noRepoUrlTest() {

        thrown.expect(Exception)
        thrown.expectMessage("ERROR - NO VALUE AVAILABLE FOR repoUrl")

        withPipeline(noRepoUrlPipeline()).execute()

    }


    private defaultPipeline() {
        return """
               @Library('piper-library-os')

               execute() {

                   pipelineExecute repoUrl: "https://test.com/myRepo.git"

               }

               return this
               """
    }

    private parameterizePipeline() {
        return """
               @Library('piper-library-os')

               execute() {

                   pipelineExecute repoUrl: "https://test.com/anotherRepo.git", branch: 'feature', path: 'path/to/Jenkinsfile', credentialsId: 'abcd1234'

               }

               return this
               """
    }

    private noRepoUrlPipeline() {
        return """
               @Library('piper-library-os')

               execute() {

                   pipelineExecute()

               }

               return this
               """
    }
}
