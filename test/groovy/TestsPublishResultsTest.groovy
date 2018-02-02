import org.junit.Before
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

    def stepUnderTest

    @Before
    void init() {
        publisherStepOptions = [:]
        archiveStepPatterns = []
        // prepare checkResultsPublish step
        stepUnderTest = loadScript('testsPublishResults.groovy').testsPublishResults
        // add handler for generic step call
        helper.registerAllowedMethod('step', [Map.class], {
            parameters -> publisherStepOptions[parameters.$class] = parameters
        })
        helper.registerAllowedMethod('junit', [Map.class], {
            parameters -> publisherStepOptions['junit'] = parameters
        })
        helper.registerAllowedMethod('jacoco', [Map.class], {
            parameters -> publisherStepOptions['jacoco'] = parameters
        })
        helper.registerAllowedMethod('cobertura', [Map.class], {
            parameters -> publisherStepOptions['cobertura'] = parameters
        })
        helper.registerAllowedMethod('archiveArtifacts', [Map.class], {
            parameters -> archiveStepPatterns.push(parameters.artifacts)
        })
    }

    @Test
    void testPublishWithDefaultSettings() throws Exception {
        stepUnderTest.call()

        // ensure nothing is published
        assertTrue('WarningsPublisher options not empty', publisherStepOptions['junit'] != null && !publisherStepOptions['junit'].active)
        assertTrue('PmdPublisher options not empty', publisherStepOptions['jacoco'] != null && !publisherStepOptions['jacoco'].active)
        assertTrue('DryPublisher options not empty', publisherStepOptions['cobertura'] != null && !publisherStepOptions['cobertura'].active)
        assertTrue('FindBugsPublisher options not empty', publisherStepOptions['PerformancePublisher'] != null && !publisherStepOptions['PerformancePublisher'].active)
    }

    @Test
    void testPublishAllWithDefaultSettings() throws Exception {
        stepUnderTest.call(junit: true, jacoco: true, cobertura: true, jmeter: false)

        assertTrue('JUnit options are empty', publisherStepOptions.junit != null)
        assertTrue('JaCoCo options are empty', publisherStepOptions.jacoco != null)
        assertTrue('Cobertura options are empty', publisherStepOptions.cobertura != null)
        //assertTrue('FindBugsPublisher options not empty', publisherStepOptions['PerformancePublisher']?.active)

        // ensure default patterns are set
        assertEquals('JUnit default pattern not set correct',
            '**/target/surefire-reports/*.xml',
            publisherStepOptions.junit.testResults)
        assertEquals('JaCoCo default pattern not set correct',
            '**/target/*.exec',
            publisherStepOptions.jacoco.execPattern)
        assertEquals('Cobertura default pattern not set correct',
            '**/target/coverage/cobertura-coverage.xml',
            publisherStepOptions.cobertura.coberturaReportFile)
        //assertEquals('CheckStylePublisher default pattern not set', '**/*.jtl', publisherStepOptions['PerformancePublisher']['pattern'])
    }
}
