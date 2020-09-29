import hudson.FilePath
import org.junit.Assert
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain

import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules
import com.sap.piper.JenkinsUtils

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

	class AnnotatedLargeTextMock {
		void writeLogTo(i, out) {}
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
		binding.setVariable('currentBuild', [result: 'UNSTABLE', rawBuild: [getLogText: { return new AnnotatedLargeTextMock() } ]])
		binding.setVariable('env', [NODE_NAME: 'anynode', WORKSPACE: '.'])
		stepRule.step.jenkinsMaterializeLog(map, body)
	}

    @Test
    void getFilePath_returnsValidFilePathObject() {
        final fileName = "mylog.txt"
        def expected = new FilePath(null, fileName)
        def script = loadScript("vars/jenkinsMaterializeLog.groovy")
        binding.setVariable('env', [NODE_NAME: 'anynode', WORKSPACE: '.'])
        def filePath = script.invokeMethod("getFilePath", fileName, new JenkinsUtilsMock())
        Assert.assertEquals(expected, filePath)
    }
}
