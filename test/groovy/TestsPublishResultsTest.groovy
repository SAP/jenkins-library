import org.junit.Before
import org.junit.Ignore
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain

import com.lesfurets.jenkins.unit.BasePipelineTest

import static org.junit.Assert.assertEquals
import static org.junit.Assert.assertTrue

import util.Rules

class TestsPublishResultsTest extends BasePipelineTest {
    Map publisherStepOptions
    List archiveStepPatterns

    @Rule
    public RuleChain ruleChain = RuleChain.outerRule(Rules.getCommonRules(this))

    def testsPublishResultsScript

    @Before
    void init() {
        publisherStepOptions = [:]
        archiveStepPatterns = []
        // prepare checkResultsPublish step
        testsPublishResultsScript = loadScript('testsPublishResults.groovy').testsPublishResults
        helper.registerAllowedMethod('junit', [Map.class], {
            parameters -> publisherStepOptions['junit'] = parameters
        })
        helper.registerAllowedMethod('jacoco', [Map.class], {
            parameters -> publisherStepOptions['jacoco'] = parameters
        })
        helper.registerAllowedMethod('cobertura', [Map.class], {
            parameters -> publisherStepOptions['cobertura'] = parameters
        })
        helper.registerAllowedMethod('perfReport', [Map.class], {
            parameters -> publisherStepOptions['jmeter'] = parameters
        })
        helper.registerAllowedMethod('archiveArtifacts', [Map.class], {
            parameters -> archiveStepPatterns.push(parameters.artifacts)
        })
    }

    @Test
    void testPublishNothingWithDefaultSettings() throws Exception {
        testsPublishResultsScript.call()

        // ensure nothing is published
        assertTrue('WarningsPublisher options not empty', publisherStepOptions.junit == null)
        assertTrue('PmdPublisher options not empty', publisherStepOptions.jacoco == null)
        assertTrue('DryPublisher options not empty', publisherStepOptions.cobertura == null)
        assertTrue('FindBugsPublisher options not empty', publisherStepOptions.jmeter == null)
    }

    @Test
    void testPublishNothingWithAllDisabled() throws Exception {
        testsPublishResultsScript.call(junit: false, jacoco: false, cobertura: false, jmeter: false)

        // ensure nothing is published
        assertTrue('WarningsPublisher options not empty', publisherStepOptions.junit == null)
        assertTrue('PmdPublisher options not empty', publisherStepOptions.jacoco == null)
        assertTrue('DryPublisher options not empty', publisherStepOptions.cobertura == null)
        assertTrue('FindBugsPublisher options not empty', publisherStepOptions.jmeter == null)
    }

    @Test
    void testPublishUnitTestsWithDefaultSettings() throws Exception {
        testsPublishResultsScript.call(junit: true)

        assertTrue('JUnit options are empty', publisherStepOptions.junit != null)
        // ensure default patterns are set
        assertEquals('JUnit default pattern not set correct',
            '**/target/surefire-reports/*.xml', publisherStepOptions.junit.testResults)
        // ensure nothing else is published
        assertTrue('JaCoCo options are not empty', publisherStepOptions.jacoco == null)
        assertTrue('Cobertura options are not empty', publisherStepOptions.cobertura == null)
        assertTrue('JMeter options are not empty', publisherStepOptions.jmeter == null)
    }

    @Test
    void testPublishCoverageWithDefaultSettings() throws Exception {
        testsPublishResultsScript.call(jacoco: true, cobertura: true)

        assertTrue('JaCoCo options are empty', publisherStepOptions.jacoco != null)
        assertTrue('Cobertura options are empty', publisherStepOptions.cobertura != null)
        assertEquals('JaCoCo default pattern not set correct',
            '**/target/*.exec', publisherStepOptions.jacoco.execPattern)
        assertEquals('Cobertura default pattern not set correct',
            '**/target/coverage/cobertura-coverage.xml', publisherStepOptions.cobertura.coberturaReportFile)
        // ensure nothing else is published
        assertTrue('JUnit options are not empty', publisherStepOptions.junit == null)
        assertTrue('JMeter options are not empty', publisherStepOptions.jmeter == null)
    }

    @Test
    void testPublishJMeterWithDefaultSettings() throws Exception {
        testsPublishResultsScript.call(jmeter: true)

        assertTrue('JMeter options are empty', publisherStepOptions.jmeter != null)
        assertEquals('JMeter default pattern not set',
            '**/*.jtl', publisherStepOptions.jmeter.sourceDataFiles)

        // ensure nothing else is published
        assertTrue('JUnit options are not empty', publisherStepOptions.junit == null)
        assertTrue('JaCoCo options are not empty', publisherStepOptions.jacoco == null)
        assertTrue('Cobertura options are not empty', publisherStepOptions.cobertura == null)
    }

    @Test
    void testPublishUnitTestsWithCustomSettings() throws Exception {
        testsPublishResultsScript.call(junit: [pattern: 'fancy/file/path', archive: true, active: true])

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
    }
}
