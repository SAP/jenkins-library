import util.BasePiperTest
import util.Rules

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.JenkinsReadYamlRule
import util.JenkinsStepRule

class PipelineExecuteTest extends BasePiperTest {
    private ExpectedException thrown = new ExpectedException().none()
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(jsr)

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
        jsr.step.call(repoUrl: "https://test.com/myRepo.git")
        assert load == "Jenkinsfile"
        assert checkoutParameters.branch == 'master'
        assert checkoutParameters.repoUrl == "https://test.com/myRepo.git"
        assert checkoutParameters.credentialsId == ''
        assert checkoutParameters.path == 'Jenkinsfile'

    }

    @Test
    void parameterizeTest() {
        jsr.step.call(repoUrl: "https://test.com/anotherRepo.git",
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

        jsr.step.call()
    }
}
