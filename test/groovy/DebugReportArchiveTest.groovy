import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.JenkinsWriteFileRule
import util.Rules

import static org.hamcrest.Matchers.containsString
import static org.junit.Assert.assertThat

class DebugReportArchiveTest extends BasePiperTest {

    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)
    private JenkinsWriteFileRule writeFileRule = new JenkinsWriteFileRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(loggingRule)
        .around(writeFileRule)
        .around(stepRule)

    @Before
    void init() {
        helper.registerAllowedMethod("libraryResource", [String.class], { path ->

            File resource = new File(new File('resources'), path)
            if (resource.exists()) {
                return resource.getText()
            }

            return ''
        })
    }

    @Test
    void testDebugReportArchive() {
        stepRule.step.debugReportArchive(
            script: nullScript,
            juStabUtils: utils,
            stageName: 'test',
            printToConsole: true
        )

        String debugReportSnippet = 'The debug log is generated with each build and should be included in every support request'

        assertThat(loggingRule.log, containsString('Successfully archived debug report'))
        assertThat(loggingRule.log, containsString(debugReportSnippet))

        assertThat(writeFileRule.files.find({ it.toString().contains('debug_log') }) as String, containsString(debugReportSnippet))
    }
}
