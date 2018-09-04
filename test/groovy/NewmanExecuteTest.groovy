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
import util.JenkinsReadYamlRule
import util.JenkinsShellCallRule
import util.JenkinsDockerExecuteRule
import util.Rules
import org.junit.rules.ExpectedException

class NewmanExecuteTest extends BasePiperTest {
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
        jsr.step.newmanExecute(
            script: nullScript,
            juStabUtils: utils,
            newmanCollection: 'testCollection',
            newmanEnvironment: 'testEnvironment',
            newmanGlobals: 'testGlobals'
        )
        // asserts
        assertThat(jscr.shell, hasItem('newman run testCollection --environment \'testEnvironment\' --globals \'testGlobals\' --reporters junit,html --reporter-junit-export target/newman/TEST-testCollection.xml --reporter-html-export target/newman/TEST-testCollection.html'))
        assertThat(jedr.dockerParams.dockerImage, is('node:8-stretch'))
        assertThat(jlr.log, containsString('[newmanExecute] Found files [testCollection]'))
        assertJobStatusSuccess()
    }

    @Test
    void testExecuteNewmanWithNoCollection() throws Exception {
        thrown.expectMessage('[newmanExecute] No collection found with pattern \'notFound.json\'')

        jsr.step.newmanExecute(
            script: nullScript,
            juStabUtils: utils,
            newmanCollection: 'notFound.json'
        )
        // asserts
        assertJobStatusFailure()
    }

    @Test
    void testExecuteNewmanFailOnError() throws Exception {
        jsr.step.newmanExecute(
            script: nullScript,
            juStabUtils: utils,
            newmanCollection: 'testCollection',
            newmanEnvironment: 'testEnvironment',
            newmanGlobals: 'testGlobals',
            dockerImage: 'testImage',
            testRepository: 'testRepo',
            failOnError: false
        )
        // asserts
        assertThat(jedr.dockerParams.dockerImage, is('testImage'))
        assertThat(gitMap.url, is('testRepo'))
        assertThat(jscr.shell, hasItem('newman run testCollection --environment \'testEnvironment\' --globals \'testGlobals\' --reporters junit,html --reporter-junit-export target/newman/TEST-testCollection.xml --reporter-html-export target/newman/TEST-testCollection.html --suppress-exit-code'))
        assertJobStatusSuccess()
    }

    @Test
    void testExecuteNewmanWithFolder() throws Exception {
        jsr.step.newmanExecute(
            script: nullScript,
            juStabUtils: utils,
            newmanRunCommand: 'run ${config.newmanCollection} --iteration-data testDataFile --reporters junit,html --reporter-junit-export target/newman/TEST-${config.newmanCollection.toString().replace(File.separatorChar,(char)\'_\').tokenize(\'.\').first()}.xml --reporter-html-export target/newman/TEST-${config.newmanCollection.toString().replace(File.separatorChar,(char)\'_\').tokenize(\'.\').first()}.html'
        )
        // asserts
        assertThat(jscr.shell, hasItem('newman run testCollectionsFolder'+File.separatorChar+'A.postman_collection.json --iteration-data testDataFile --reporters junit,html --reporter-junit-export target/newman/TEST-testCollectionsFolder_A.xml --reporter-html-export target/newman/TEST-testCollectionsFolder_A.html'))
        assertThat(jscr.shell, hasItem('newman run testCollectionsFolder'+File.separatorChar+'B.postman_collection.json --iteration-data testDataFile --reporters junit,html --reporter-junit-export target/newman/TEST-testCollectionsFolder_B.xml --reporter-html-export target/newman/TEST-testCollectionsFolder_B.html'))
        assertJobStatusSuccess()
    }
}
