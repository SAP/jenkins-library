
import org.junit.Before
import org.junit.Rule
import org.junit.Test

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertTrue

import util.JenkinsSetupRule

class MavenExecuteTest extends PiperTestBase {

    Map dockerParameters
    List shellCalls

    @Rule
    public JenkinsSetupRule jsr = new JenkinsSetupRule(this)

    @Before
    void init() {

        shellCalls = []
        dockerParameters = [:]

        helper.registerAllowedMethod("dockerExecute", [Map.class, Closure.class],
            { parameters, closure ->
                dockerParameters = parameters
                closure()
            })
        helper.registerAllowedMethod('sh', [String], { s -> shellCalls.add(s) })
    }

    @Test
    void testExecuteBasicMavenCommand() throws Exception {
        def script = loadScript("test/resources/pipelines/mavenExecuteTest/executeBasicMavenCommand.groovy")
        script.execute()
        assertEquals('maven:3.5-jdk-7', dockerParameters.dockerImage)
        assertTrue(shellCalls.contains('mvn clean install'))
    }

    @Test
    void testExecuteMavenCommandWithParameter() throws Exception {
        def script = loadScript("test/resources/pipelines/mavenExecuteTest/executeMavenCommandWithParameters.groovy")
        script.execute()
        assertEquals('maven:3.5-jdk-8-alpine', dockerParameters.dockerImage)
        String mvnCommand = "mvn --global-settings 'globalSettingsFile.xml' -Dmaven.repo.local='m2Path' --settings 'projectSettingsFile.xml' --file 'pom.xml' -o clean install -Dmaven.tests.skip=true"
        assertTrue(shellCalls.contains(mvnCommand))
    }
}
