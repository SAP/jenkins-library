import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsEnvironmentRule
import util.JenkinsFileExistsRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.Rules

import static org.junit.Assert.*

public class DeleteFileTest extends BasePiperTest {

    private JenkinsReadYamlRule readYamlRule = new JenkinsReadYamlRule(this)
    private JenkinsFileExistsRule fileExistsRule = new JenkinsFileExistsRule(this, [])
    private JenkinsStepRule script = new JenkinsStepRule(this)
    private JenkinsEnvironmentRule environmentRule = new JenkinsEnvironmentRule(this)
    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule shellCallRule = new JenkinsShellCallRule(this)
    private ExpectedException expectedExceptionRule = ExpectedException.none()


    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(shellCallRule)
        .around(readYamlRule)
        .around(fileExistsRule)
        .around(environmentRule)
        .around(loggingRule)
        .around(script)
        .around(expectedExceptionRule)

    @Before
    public void setup() {

    }

    @Test
    public void deleteFile() throws Exception {
        String filePath = "test.file"
        String command = "rm '${filePath}'"
        Integer statusSuccess = 0

        fileExistsRule.registerExistingFile(filePath)
        shellCallRule.setReturnValue(command, statusSuccess)
        loggingRule.expect("[DeleteFile] Successfully deleted file '${filePath}'.")

        // execute step
        script.step.deleteFile path: filePath, script: nullScript

        assertTrue(fileExistsRule.queriedFiles.contains(filePath))
        assertTrue(shellCallRule.shell.contains(command))
    }

    @Test
    public void deleteFile_failsSilently_IfFileDoesNotExist() throws Exception {
        String filePath = "test.file"

        // execute step
        script.step.deleteFile path: filePath, script: nullScript

        assertTrue(fileExistsRule.queriedFiles.contains(filePath))
        assertFalse(loggingRule.log.contains("[DeleteFile]"))
    }

    @Test
    public void deleteFile_Throws_OnInvalidPath() throws Exception {
        String filePath = "test.file"
        String command = "rm ${filePath}"
        Integer statusSuccess = 0

        fileExistsRule.registerExistingFile(filePath)
        shellCallRule.setReturnValue(command, statusSuccess)

        expectedExceptionRule.expect(hudson.AbortException)
        expectedExceptionRule.expectMessage("[DeleteFile] File path must not be null or empty.")

        // execute step
        script.step.deleteFile path: null, script: nullScript

        assertFalse(fileExistsRule.queriedFiles.contains(filePath))
        assertFalse(shellCallRule.shell.contains(command))
    }

    @Test
    public void deleteFile_Throws_IfFileExistsButCouldNotBeDeleted() throws Exception {
        String filePath = "test.file"
        String command = "rm '${filePath}'"
        Integer statusError = 1

        fileExistsRule.registerExistingFile(filePath)
        shellCallRule.setReturnValue(command, statusError)

        expectedExceptionRule.expect(hudson.AbortException)
        expectedExceptionRule.expectMessage("[DeleteFile] Could not delete file '${filePath}'. Check file permissions.")

        // execute step
        script.step.deleteFile path: filePath, script: nullScript

        assertTrue(fileExistsRule.queriedFiles.contains(filePath))
        assertTrue(shellCallRule.shell.contains(command))
    }
}
