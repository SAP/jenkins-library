
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain

import util.JenkinsDockerExecuteRule
import util.JenkinsStepRule

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertTrue

import util.BasePiperTest
import util.JenkinsShellCallRule
import util.Rules

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

        assert jscr.shell[0] == 'mvn clean install'
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
        String mvnCommand = "mvn --global-settings 'globalSettingsFile.xml' -Dmaven.repo.local='m2Path' --settings 'projectSettingsFile.xml' --file 'pom.xml' -o clean install -Dmaven.tests.skip=true"
        assertTrue(jscr.shell.contains(mvnCommand))
    }

    @Test
    void testMavenCommandForwardsDockerOptions() throws Exception {

        jsr.step.mavenExecute(script: nullScript, goals: 'clean install')
        assertEquals('maven:3.5-jdk-7', jder.dockerParams.dockerImage)

        assert jscr.shell[0] == 'mvn clean install'
    }
}
