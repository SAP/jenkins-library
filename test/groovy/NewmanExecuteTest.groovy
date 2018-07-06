import org.junit.Before
import org.junit.Test
import org.junit.Rule
import org.junit.rules.RuleChain

import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.containsString

import static org.junit.Assert.assertThat

import util.BasePiperTest
import util.JenkinsStepRule
import util.JenkinsLoggingRule
import util.JenkinsShellCallRule
import util.JenkinsDockerExecuteRule
import util.Rules

class NewmanExecuteTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)
    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsShellCallRule jscr = new JenkinsShellCallRule(this)
    private JenkinsDockerExecuteRule jedr = new JenkinsDockerExecuteRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(jedr)
        .around(jscr)
        .around(jlr)
        .around(jsr) // needs to be activated after jedr, otherwise executeDocker is not mocked

    def testRepository

    @Before
    void init() throws Exception {
        helper.registerAllowedMethod('git', [String.class], {s ->
            testRepository = s
        })
        helper.registerAllowedMethod("findFiles", [Map.class], { map ->
            def files
            if(map.glob == '**/*.postman_collection.json')
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
        jsr.step.newmanExecute(
            script: nullScript,
            newmanCollection: 'testCollection',
            newmanEnvironment: 'testEnvironment',
            newmanGlobals: 'testGlobals'
        )
        // asserts
        assertThat(jscr.shell, hasItem('newman run testCollection --environment \'testEnvironment\' --globals \'testGlobals\' --reporters junit,html --reporter-junit-export target/newman/TEST-testCollection.xml --reporter-html-export target/newman/TEST-testCollection.html'))
        assertThat(jedr.dockerParams.dockerImage, is('node:8-stretch'))
        assertJobStatusSuccess()
    }

    @Test
    void testExecuteNewmanFailOnError() throws Exception {
        jsr.step.newmanExecute(
            script: nullScript,
            newmanCollection: 'testCollection',
            newmanEnvironment: 'testEnvironment',
            newmanGlobals: 'testGlobals',
            dockerImage: 'testImage',
            testRepository: 'testRepo',
            failOnError: false
        )
        // asserts
        assertThat(jedr.dockerParams.dockerImage, is('testImage'))
        assertThat(testRepository, is('testRepo'))
        assertThat(jscr.shell, hasItem('newman run testCollection --environment \'testEnvironment\' --globals \'testGlobals\' --reporters junit,html --reporter-junit-export target/newman/TEST-testCollection.xml --reporter-html-export target/newman/TEST-testCollection.html --suppress-exit-code'))
        assertJobStatusSuccess()
    }

    @Test
    void testExecuteNewmanWithFolder() throws Exception {
        jsr.step.newmanExecute(
            script: nullScript,
            newmanRunCommand: 'run ${config.newmanCollection} --iteration-data testDataFile --reporters junit,html --reporter-junit-export target/newman/TEST-${config.newmanCollection.toString().replaceAll(\'/\',\'_\').tokenize(\'.\').first()}.xml --reporter-html-export target/newman/TEST-${config.newmanCollection.toString().replaceAll(\'/\',\'_\').tokenize(\'.\').first()}.html'
        )
        // asserts
        assertThat(jscr.shell, hasItem('newman run testCollectionsFolder/A.postman_collection.json --iteration-data testDataFile --reporters junit,html --reporter-junit-export target/newman/TEST-testCollectionsFolder_A.xml --reporter-html-export target/newman/TEST-testCollectionsFolder_A.html'))
        assertThat(jscr.shell, hasItem('newman run testCollectionsFolder/B.postman_collection.json --iteration-data testDataFile --reporters junit,html --reporter-junit-export target/newman/TEST-testCollectionsFolder_B.xml --reporter-html-export target/newman/TEST-testCollectionsFolder_B.html'))
        assertJobStatusSuccess()
    }
}
