
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertTrue

import util.JenkinsConfigRule
import util.JenkinsSetupRule
import util.JenkinsShellCallRule

class MavenExecuteTest extends PiperTestBase {

    Map dockerParameters

    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)

    @Rule
    public RuleChain ruleChain = RuleChain.outerRule(new JenkinsSetupRule(this))
                                              .around(jscr)
                                              .around(new JenkinsConfigRule(this))

    @Before
    void init() {

        dockerParameters = [:]

        helper.registerAllowedMethod("dockerExecute", [Map.class, Closure.class],
            { parameters, closure ->
                dockerParameters = parameters
                closure()
            })
    }

    @Test
    void testExecuteBasicMavenCommand() throws Exception {
        def script = loadScript("test/resources/pipelines/mavenExecuteTest/executeBasicMavenCommand.groovy")
        script.execute()
        assertEquals('maven:3.5-jdk-7', dockerParameters.dockerImage)

        assert jscr.shell[0] == 'mvn clean install'
    }

    @Test
    void testExecuteMavenCommandWithParameter() throws Exception {
        def script = loadScript("test/resources/pipelines/mavenExecuteTest/executeMavenCommandWithParameters.groovy")
        script.execute()
        assertEquals('maven:3.5-jdk-8-alpine', dockerParameters.dockerImage)
        String mvnCommand = "mvn --global-settings 'globalSettingsFile.xml' -Dmaven.repo.local='m2Path' --settings 'projectSettingsFile.xml' --file 'pom.xml' -o clean install -Dmaven.tests.skip=true"
        assertTrue(jscr.shell.contains(mvnCommand))
    }
}
