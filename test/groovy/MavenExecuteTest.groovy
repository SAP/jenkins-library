import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsDockerExecuteRule
import util.JenkinsShellCallRule
import util.JenkinsStepRule
import util.Rules

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.not
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue


class MavenExecuteTest extends BasePiperTest {

    Map dockerParameters

    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsDockerExecuteRule jder = new JenkinsDockerExecuteRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(jder)
        .around(jscr)
        .around(jsr)

    @Test
    void testExecuteBasicMavenCommand() throws Exception {

        jsr.step.mavenExecute(script: nullScript, goals: 'clean install')
        assertEquals('maven:3.5-jdk-7', jder.dockerParams.dockerImage)

        assert jscr.shell[0] == 'mvn --batch-mode -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn clean install'
    }

    @Test
    void testExecuteBasicMavenCommandWithDownloadLogsEnabled() throws Exception {

        jsr.step.mavenExecute(script: nullScript, goals: 'clean install', logSuccessfulMavenTransfers: true)
        assertEquals('maven:3.5-jdk-7', jder.dockerParams.dockerImage)

        assert jscr.shell[0] == 'mvn --batch-mode clean install'
    }

    @Test
    void testExecuteMavenCommandWithParameter() throws Exception {

        jsr.step.mavenExecute(
            script: nullScript,
            dockerImage: 'maven:3.5-jdk-8-alpine',
            goals: 'clean install',
            globalSettingsFile: 'globalSettingsFile.xml',
            projectSettingsFile: 'projectSettingsFile.xml',
            pomPath: 'pom.xml',
            flags: '-o',
            m2Path: 'm2Path',
            defines: '-Dmaven.tests.skip=true')
        assertEquals('maven:3.5-jdk-8-alpine', jder.dockerParams.dockerImage)
        String mvnCommand = "mvn --global-settings 'globalSettingsFile.xml' -Dmaven.repo.local='m2Path' --settings 'projectSettingsFile.xml' --file 'pom.xml' -o --batch-mode -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn clean install -Dmaven.tests.skip=true"
        assertTrue(jscr.shell.contains(mvnCommand))
    }

    @Test
    void testMavenCommandForwardsDockerOptions() throws Exception {
        jsr.step.mavenExecute(script: nullScript, goals: 'clean install')
        assertEquals('maven:3.5-jdk-7', jder.dockerParams.dockerImage)

        assertEquals('mvn --batch-mode -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn clean install', jscr.shell[0])
    }

    @Test
    void testMavenCommandWithShortBatchModeFlag() throws Exception {
        jsr.step.mavenExecute(script: nullScript, goals: 'clean install', flags: '-B')
        assertEquals('maven:3.5-jdk-7', jder.dockerParams.dockerImage)

        assertEquals('mvn -B -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn clean install', jscr.shell[0])
    }

    @Test
    void testMavenCommandWithFalsePositiveMinusBFlag() throws Exception {
        jsr.step.mavenExecute(script: nullScript, goals: 'clean install', flags: '-Blah')
        assertEquals('maven:3.5-jdk-7', jder.dockerParams.dockerImage)

        assertThat(jscr.shell[0],
            allOf(containsString('-Blah'),
                  containsString('--batch-mode')))
    }

    @Test
    void testMavenCommandWithBatchModeMultiline() throws Exception {
        jsr.step.mavenExecute(script: nullScript, goals: 'clean install', flags: ('''-B\\
                                                                                    |--show-version''' as CharSequence).stripMargin())
        assertThat(jscr.shell[0], not(containsString('--batch-mode')))
    }
}
