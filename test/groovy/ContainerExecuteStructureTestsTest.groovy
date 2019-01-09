import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class ContainerExecuteStructureTestsTest extends BasePiperTest {
    private ExpectedException thrown = ExpectedException.none()
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsDockerExecuteRule jedr = new JenkinsDockerExecuteRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(jedr)
        .around(jscr)
        .around(jlr)
        .around(jsr) // needs to be activated after jedr, otherwise executeDocker is not mocked

    def gitMap

    @Before
    void init() throws Exception {
        helper.registerAllowedMethod('stash', [String.class], null)
        helper.registerAllowedMethod('git', [Map.class], {m ->
            gitMap = m
        })
        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            def files
            if(map.glob == 'notFound.json')
                files = []
            else if(map.glob == '**/*.postman_collection.json')
                files = [
                    new File("testCollectionsFolder/A.postman_collection.json"),
                    new File("testCollectionsFolder/B.postman_collection.json")
                ]
            else
                files = [new File(map.glob)]
            return files.toArray()
        })
    }

    @Test
    void testExecuteNewmanDefault() throws Exception {
        jsr.step.containerExecuteStructureTests(
            script: nullScript,
            juStabUtils: utils,
        )
        // asserts
        assertThat(jscr.shell, hasItem('xxx'))
        assertThat(jedr.dockerParams.dockerImage, is('node:8-stretch'))
        assertThat(jlr.log, containsString('[newmanExecute] Found files [testCollection]'))
        assertJobStatusSuccess()
    }
}
