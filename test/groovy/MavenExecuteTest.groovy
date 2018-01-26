
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain

import com.lesfurets.jenkins.unit.BasePipelineTest

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertTrue

import util.JenkinsShellCallRule
import util.Rules

class MavenExecuteTest extends BasePipelineTest {

    Map dockerParameters

    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
                                      .around(jscr)

    def mavenExecuteScript
    def cpe

    @Before
    void init() {

        dockerParameters = [:]

        helper.registerAllowedMethod("dockerExecute", [Map.class, Closure.class],
            { parameters, closure ->
                dockerParameters = parameters
                closure()
            })

        mavenExecuteScript = loadScript("mavenExecute.groovy").mavenExecute
        cpe = loadScript('commonPipelineEnvironment.groovy').commonPipelineEnvironment
    }

    @Test
    void testExecuteBasicMavenCommand() throws Exception {

        mavenExecuteScript.call(script: [commonPipelineEnvironment: cpe], goals: 'clean install')
        assertEquals('maven:3.5-jdk-7', dockerParameters.dockerImage)

        assert jscr.shell[0] == 'mvn clean install'
    }

    @Test
    void testExecuteMavenCommandWithParameter() throws Exception {

        mavenExecuteScript.call(
            script: [commonPipelineEnvironment: cpe],
            dockerImage: 'maven:3.5-jdk-8-alpine',
            goals: 'clean install',
            globalSettingsFile: 'globalSettingsFile.xml',
            projectSettingsFile: 'projectSettingsFile.xml',
            pomPath: 'pom.xml',
            flags: '-o',
            m2Path: 'm2Path',
            defines: '-Dmaven.tests.skip=true')
        assertEquals('maven:3.5-jdk-8-alpine', dockerParameters.dockerImage)
        String mvnCommand = "mvn --global-settings 'globalSettingsFile.xml' -Dmaven.repo.local='m2Path' --settings 'projectSettingsFile.xml' --file 'pom.xml' -o clean install -Dmaven.tests.skip=true"
        assertTrue(jscr.shell.contains(mvnCommand))
    }
}
