import util.JenkinsSetupRule

import com.lesfurets.jenkins.unit.BasePipelineTest

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.JenkinsReadYamlRule

class PipelineExecuteTest extends BasePipelineTest {

    private ExpectedException thrown = new ExpectedException().none()

    @Rule
    public RuleChain ruleChain = RuleChain.outerRule(thrown)
                                              .around(new JenkinsSetupRule(this))
                                              .around(new JenkinsReadYamlRule(this))

    def pipelinePath
    def checkoutParameters = [:]
    def load

    def pipelineExecuteScript

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

        pipelineExecuteScript = loadScript("pipelineExecute.groovy").pipelineExecute
    }


    @Test
    void straightForwardTest() {

        pipelineExecuteScript.call(repoUrl: "https://test.com/myRepo.git")
        assert load == "Jenkinsfile"
        assert checkoutParameters.branch == 'master'
        assert checkoutParameters.repoUrl == "https://test.com/myRepo.git"
        assert checkoutParameters.credentialsId == ''
        assert checkoutParameters.path == 'Jenkinsfile'

    }

    @Test
    void parameterizeTest() {

        pipelineExecuteScript.call(repoUrl: "https://test.com/anotherRepo.git",
                             branch: 'feature',
                             path: 'path/to/Jenkinsfile',
                             credentialsId: 'abcd1234')

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

        pipelineExecuteScript.call()
    }
}
