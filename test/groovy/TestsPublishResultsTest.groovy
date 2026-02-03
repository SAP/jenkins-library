import com.sap.piper.BashUtils
import org.junit.After
import org.junit.Before
import org.junit.Ignore
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import org.junit.rules.ExpectedException

import util.BasePiperTest
import util.JenkinsReadYamlRule
import util.JenkinsStepRule

import static org.hamcrest.Matchers.not
import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue
import static org.hamcrest.Matchers.containsString

import com.sap.piper.Utils

import util.Rules
import minimatch.Minimatch

class TestsPublishResultsTest extends BasePiperTest {
    Map publisherStepOptions
    List archiveStepPatterns

    private ExpectedException thrown = ExpectedException.none()
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(thrown)
        .around(stepRule)

    @Before
    void init() {
        publisherStepOptions = [:]
        archiveStepPatterns = []
        // prepare checkResultsPublish step
        helper.registerAllowedMethod('junit', [Map.class], {
            parameters -> publisherStepOptions['junit'] = parameters
        })
        helper.registerAllowedMethod('jacoco', [Map.class], {
            parameters -> publisherStepOptions['jacoco'] = parameters
        })
        helper.registerAllowedMethod('recordCoverage', [Map.class], {
            parameters -> publisherStepOptions['cobertura'] = parameters
        })
        helper.registerAllowedMethod('perfReport', [Map.class], {
            parameters -> publisherStepOptions['jmeter'] = parameters
        })
        helper.registerAllowedMethod('cucumber', [Map.class], {
            parameters -> publisherStepOptions['cucumber'] = parameters
        })
        helper.registerAllowedMethod('publishHTML', [Map.class], {
            parameters -> publisherStepOptions['htmlPublisher'] = parameters
        })
        helper.registerAllowedMethod('archiveArtifacts', [Map.class], {
            parameters -> archiveStepPatterns.push(parameters.artifacts)
        })

        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
    }

    @Test
    void testPublishNothingWithDefaultSettings() throws Exception {
        stepRule.step.testsPublishResults(script: nullScript)

        // ensure nothing is published
        assertTrue('WarningsPublisher options not empty', publisherStepOptions.junit == null)
        assertTrue('PmdPublisher options not empty', publisherStepOptions.jacoco == null)
        assertTrue('DryPublisher options not empty', publisherStepOptions.cobertura == null)
        assertTrue('FindBugsPublisher options not empty', publisherStepOptions.jmeter == null)
        assertTrue('Cucumber options not empty', publisherStepOptions.cucumber == null)
        assertTrue('HtmlPublisher options not empty', publisherStepOptions.htmlPublisher == null)
    }

    @Test
    void testPublishNothingWithAllDisabled() throws Exception {
        stepRule.step.testsPublishResults(script: nullScript, junit: false, jacoco: false, cobertura: false, jmeter: false, cucumber: false, htmlPublisher: false)

        // ensure nothing is published
        assertTrue('WarningsPublisher options not empty', publisherStepOptions.junit == null)
        assertTrue('PmdPublisher options not empty', publisherStepOptions.jacoco == null)
        assertTrue('DryPublisher options not empty', publisherStepOptions.cobertura == null)
        assertTrue('FindBugsPublisher options not empty', publisherStepOptions.jmeter == null)
        assertTrue('Cucumber options not empty', publisherStepOptions.cucumber == null)
        assertTrue('HtmlPublisher options not empty', publisherStepOptions.htmlPublisher == null)
    }

    @Test
    void testPublishUnitTestsWithDefaultSettings() throws Exception {
        stepRule.step.testsPublishResults(script: nullScript, junit: true)

        assertTrue('JUnit options are empty', publisherStepOptions.junit != null)
        // ensure default patterns are set
        assertEquals('JUnit default pattern not set correct',
            '**/TEST-*.xml', publisherStepOptions.junit.testResults)
        // ensure nothing else is published
        assertTrue('JaCoCo options are not empty', publisherStepOptions.jacoco == null)
        assertTrue('Cobertura options are not empty', publisherStepOptions.cobertura == null)
        assertTrue('JMeter options are not empty', publisherStepOptions.jmeter == null)
        assertTrue('Cucumber options not empty', publisherStepOptions.cucumber == null)
        assertTrue('HtmlPublisher options not empty', publisherStepOptions.htmlPublisher == null)
    }

    @Test
    void testPublishCoverageWithDefaultSettings() throws Exception {
        stepRule.step.testsPublishResults(script: nullScript, jacoco: true, cobertura: true)

        assertTrue('JaCoCo options are empty', publisherStepOptions.jacoco != null)
        assertEquals('JaCoCo default pattern not set correct',
            '**/target/*.exec', publisherStepOptions.jacoco.execPattern)
        // ensure nothing else is published
        assertTrue('JUnit options are not empty', publisherStepOptions.junit == null)
        assertTrue('JMeter options are not empty', publisherStepOptions.jmeter == null)
        assertTrue('Cucumber options not empty', publisherStepOptions.cucumber == null)
        assertTrue('HtmlPublisher options not empty', publisherStepOptions.htmlPublisher == null)

//         assertTrue('Cobertura options are empty', publisherStepOptions.cobertura != null)
//         assertTrue('Cobertura default pattern is empty', publisherStepOptions.cobertura.coberturaReportFile != null)
//         String sampleCoberturaPathForJava = 'my/workspace/my/project/target/coverage/cobertura-coverage.xml'
//         assertTrue('Cobertura default pattern does not match files at target/coverage/cobertura-coverage.xml for Java projects',
//             Minimatch.minimatch(sampleCoberturaPathForJava, publisherStepOptions.cobertura.coberturaReportFile))
//         String sampleCoberturaPathForKarma = 'my/workspace/my/project/target/coverage/Chrome 78.0.3904 (Mac OS X 10.14.6)/cobertura-coverage.xml'
//         assertTrue('Cobertura default pattern does not match files at target/coverage/<browser>/cobertura-coverage.xml for UI5 projects',
//             Minimatch.minimatch(sampleCoberturaPathForKarma, publisherStepOptions.cobertura.coberturaReportFile))
    }

    @Test
    void testPublishJMeterWithDefaultSettings() throws Exception {
        stepRule.step.testsPublishResults(script: nullScript, jmeter: true)

        assertTrue('JMeter options are empty', publisherStepOptions.jmeter != null)
        assertEquals('JMeter default pattern not set',
            '**/*.jtl', publisherStepOptions.jmeter.sourceDataFiles)

        // ensure nothing else is published
        assertTrue('JUnit options are not empty', publisherStepOptions.junit == null)
        assertTrue('JaCoCo options are not empty', publisherStepOptions.jacoco == null)
        assertTrue('Cobertura options are not empty', publisherStepOptions.cobertura == null)
        assertTrue('Cucumber options not empty', publisherStepOptions.cucumber == null)
        assertTrue('HtmlPublisher options not empty', publisherStepOptions.htmlPublisher == null)
    }

    @Test
    void testPublishUnitTestsWithCustomSettings() throws Exception {
        stepRule.step.testsPublishResults(script: nullScript, junit: [pattern: 'fancy/file/path', archive: true, active: true])

        assertTrue('JUnit options are empty', publisherStepOptions.junit != null)
        // ensure default patterns are set
        assertEquals('JUnit pattern not set correct',
            'fancy/file/path', publisherStepOptions.junit.testResults)
        assertEquals('JUnit default pattern not set correct',
            'fancy/file/path', publisherStepOptions.junit.testResults)
        // ensure nothing else is published
        assertTrue('JaCoCo options are not empty', publisherStepOptions.jacoco == null)
        assertTrue('Cobertura options are not empty', publisherStepOptions.cobertura == null)
        assertTrue('JMeter options are not empty', publisherStepOptions.jmeter == null)
        assertTrue('Cucumber options not empty', publisherStepOptions.cucumber == null)
        assertTrue('HtmlPublisher options not empty', publisherStepOptions.htmlPublisher == null)
    }

    @Test
    void testPublishCucumberResults() throws Exception {
        stepRule.step.testsPublishResults(script: nullScript, cucumber: [pattern: 'fancy/file/path', archive: true, active: true])

        assertTrue('Cucumber options are empty', publisherStepOptions.cucumber != null)
        assertEquals('Cucumber pattern not set correct',
            'fancy/file/path', publisherStepOptions.cucumber.testResults)

        assertTrue('JUnit options are not empty', publisherStepOptions.junit == null)
        assertTrue('JaCoCo options are not empty', publisherStepOptions.jacoco == null)
        assertTrue('Cobertura options are not empty', publisherStepOptions.cobertura == null)
        assertTrue('JMeter options are not empty', publisherStepOptions.jmeter == null)
        assertTrue('HtmlPublisher options not empty', publisherStepOptions.htmlPublisher == null)
    }

    @Test
    void testPublishHtmlResults() throws Exception {
        stepRule.step.testsPublishResults(script: nullScript, htmlPublisher: [pattern: 'fancy/file/path', active: true])

        assertTrue('HtmlPublisher options are empty', publisherStepOptions.htmlPublisher != null)
        assertEquals('HtmlPublisher pattern not set correct',
            'fancy/file/path', publisherStepOptions.htmlPublisher.target.reportFiles)

        assertTrue('JUnit options are not empty', publisherStepOptions.junit == null)
        assertTrue('JaCoCo options are not empty', publisherStepOptions.jacoco == null)
        assertTrue('Cobertura options are not empty', publisherStepOptions.cobertura == null)
        assertTrue('JMeter options are not empty', publisherStepOptions.jmeter == null)
        assertTrue('Cucumber options not empty', publisherStepOptions.cucumber == null)
    }

    @Test
    void testBuildResultStatus() throws Exception {
        stepRule.step.testsPublishResults(script: nullScript)
        assertJobStatusSuccess()
    }

    @Test
    void testBuildWithTestFailuresAndWithoutFailOnError() throws Exception {
        nullScript.currentBuild.getRawBuild = {
            return [getAction: { type ->
                return [getFailCount: {
                    return 6
                }]
            }]
        }

        stepRule.step.testsPublishResults(script: nullScript)
        assertJobStatusSuccess()
    }

    @Test
    void testBuildWithTestFailuresAndWithFailOnError() throws Exception {
        nullScript.currentBuild.getRawBuild = {
            return [getAction: { type ->
                return [getFailCount: {
                    return 6
                }]
            }]
        }

        thrown.expect(hudson.AbortException)
        thrown.expectMessage('[testsPublishResults] Some tests failed!')

        stepRule.step.testsPublishResults(script: nullScript, failOnError: true)
    }

    @Test
    void testPublishUnitTestsWithUpdateResultsDoesNotAllowCommandExecution() throws Exception {
        def injectString = "' -exec touch {} ; rm -rf / # –"
        helper.registerAllowedMethod('sh', [String], { String cmd ->
            assertThat(cmd, containsString(BashUtils.quoteAndEscape(injectString)))
        })

        stepRule.step.testsPublishResults(script: nullScript, junit: [pattern: injectString, archive: true, active: true, updateResults: true])


    }
}
