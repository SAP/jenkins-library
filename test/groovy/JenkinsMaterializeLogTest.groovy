import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import com.sap.piper.GitUtils

import hudson.AbortException
import util.BasePiperTest
import util.JenkinsDockerExecuteRule
import util.JenkinsEnvironmentRule
import util.JenkinsLoggingRule
import util.JenkinsReadMavenPomRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules
import com.sap.piper.JenkinsUtils
import jenkins.model.Jenkins

class JenkinsMaterializeLogTest extends BasePiperTest {

	private ExpectedException thrown = ExpectedException.none()
	private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
	private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)
	private JenkinsStepRule stepRule = new JenkinsStepRule(this)

	class JenkinsUtilsMock extends JenkinsUtils {
		def getInstance() {
			def map = [getComputer:{return null}];
			return map
		}
	}

	@Rule
	public RuleChain ruleChain = Rules
	.getCommonRules(this)
	.around(new JenkinsReadYamlRule(this))
	.around(thrown)
	.around(loggingRule)
	.around(writeFileRule)
	.around(stepRule)

	@Test
	void testMaterializeLog() {
		def map = [script: nullScript, jenkinsUtilsStub: new JenkinsUtilsMock()]
		def body = { name -> def msg = "hello " + name }
		binding.setVariable('currentBuild', [result: 'UNSTABLE', rawBuild: [getLogInputStream: {return new StringBufferInputStream("this is the input")}]])
		binding.setVariable('env', [NODE_NAME: 'anynode', WORKSPACE: '.'])
		stepRule.step.jenkinsMaterializeLog(map, body)
	}
}
