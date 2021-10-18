import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import org.junit.Ignore

import com.sap.piper.Utils

import util.BasePiperTest

import static org.hamcrest.Matchers.hasItem
import static org.hamcrest.Matchers.containsInAnyOrder
import static org.hamcrest.Matchers.empty
import static org.hamcrest.Matchers.not
import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.hasKey
import static org.hamcrest.Matchers.hasSize
import static org.hamcrest.Matchers.hasEntry
import static org.junit.Assert.assertThat

import util.Rules
import util.JenkinsReadYamlRule
import util.JenkinsStepRule


class ChecksPublishResultsTest extends BasePiperTest {
    Map publisherStepOptions
    List archiveStepPatterns
    List invokedReportingTools

    private JenkinsStepRule stepRule = new JenkinsStepRule(this)

    @Rule
    public RuleChain ruleChain = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(stepRule)

    @Before
    void init() {
        publisherStepOptions = [:]
        archiveStepPatterns = []
        invokedReportingTools = []

        // add handler for generic step call
        helper.registerAllowedMethod("recordIssues", [Map.class], {
            parameters ->
            if(parameters.tools[0] in Map && parameters.tools[0].containsKey('publisher')) {
              publisherStepOptions[parameters.tools[0].publisher] = parameters;
            }
        })
        helper.registerAllowedMethod("pmdParser", [Map.class], {
            parameters ->
                invokedReportingTools << "pmdParser";
                return parameters.plus([publisher: "PmdPublisher"])
        })
        helper.registerAllowedMethod("cpd", [Map.class], {
            parameters ->
                invokedReportingTools << "cpd";
                return parameters.plus([publisher: "DryPublisher"])
        })
        helper.registerAllowedMethod("findBugs", [Map.class], {
            parameters ->
                invokedReportingTools << "findBugs";
                return parameters.plus([publisher: "FindBugsPublisher"])
        })
        helper.registerAllowedMethod("checkStyle", [Map.class], {
            parameters ->
                invokedReportingTools << "checkStyle";
                return parameters.plus([publisher: "CheckStylePublisher"])
        })
        helper.registerAllowedMethod("esLint", [Map.class], {
            parameters ->
                invokedReportingTools << "esLint";
                return parameters.plus([publisher: "ESLintPublisher"])
        })
        helper.registerAllowedMethod("pyLint", [Map.class], {
            parameters ->
                invokedReportingTools << "pyLint";
                return parameters.plus([publisher: "PyLintPublisher"])
        })
        helper.registerAllowedMethod("taskScanner", [Map.class], {
            parameters ->
                invokedReportingTools << "taskScanner";
                return parameters.plus([publisher: "TaskPublisher"])
        })
        helper.registerAllowedMethod("archiveArtifacts", [Map.class], {
            parameters -> archiveStepPatterns.push(parameters.artifacts)
        })
        Utils.metaClass.echo = { def m -> }
    }

    @After
    public void tearDown() {
        Utils.metaClass = null
    }

    @Test
    void testPublishWithDefaultSettings() throws Exception {
        // test
        stepRule.step.checksPublishResults(script: nullScript)
        // assert
        // ensure nothing is published
        assertThat(publisherStepOptions, not(hasKey('PmdPublisher')))
        assertThat(publisherStepOptions, not(hasKey('DryPublisher')))
        assertThat(publisherStepOptions, not(hasKey('FindBugsPublisher')))
        assertThat(publisherStepOptions, not(hasKey('CheckStylePublisher')))
        assertThat(publisherStepOptions, not(hasKey('ESLintPublisher')))
        assertThat(publisherStepOptions, not(hasKey('PyLintPublisher')))
        assertThat(publisherStepOptions, not(hasKey('TaskPublisher')))

        assertThat(invokedReportingTools, is(empty()))
    }

    @Test
    void testPublishForJavaWithDefaultSettings() throws Exception {
        // test
        stepRule.step.checksPublishResults(script: nullScript, pmd: true, cpd: true, findbugs: true, checkstyle: true)
        // assert
        assertThat(publisherStepOptions, hasKey('PmdPublisher'))
        assertThat(publisherStepOptions, hasKey('DryPublisher'))
        assertThat(publisherStepOptions, hasKey('FindBugsPublisher'))
        assertThat(publisherStepOptions, hasKey('CheckStylePublisher'))
        // ensure default patterns are set
        assertThat(publisherStepOptions['PmdPublisher'], hasKey('tools'))
        assertThat(publisherStepOptions['PmdPublisher']['tools'], hasItem(hasEntry('pattern', '**/target/pmd.xml')))
        assertThat(publisherStepOptions['DryPublisher'], hasKey('tools'))
        assertThat(publisherStepOptions['DryPublisher']['tools'], hasItem(hasEntry('pattern', '**/target/cpd.xml')))
        assertThat(publisherStepOptions['FindBugsPublisher'], hasKey('tools'))
        assertThat(publisherStepOptions['FindBugsPublisher']['tools'], hasItem(hasEntry('pattern', '**/target/findbugsXml.xml, **/target/findbugs.xml')))
        assertThat(publisherStepOptions['CheckStylePublisher'], hasKey('tools'))
        assertThat(publisherStepOptions['CheckStylePublisher']['tools'], hasItem(hasEntry('pattern', '**/target/checkstyle-result.xml')))
        // ensure nothing else is published
        assertThat(publisherStepOptions, not(hasKey('ESLintPublisher')))
        assertThat(publisherStepOptions, not(hasKey('PyLintPublisher')))
        assertThat(publisherStepOptions, not(hasKey('TaskPublisher')))
    }

    @Test
    void testPublishForJavaScriptWithDefaultSettings() throws Exception {
        // test
        stepRule.step.checksPublishResults(script: nullScript, eslint: true)
        // assert
        assertThat(publisherStepOptions, hasKey('ESLintPublisher'))
        // ensure correct parser is set set
        assertThat(publisherStepOptions['ESLintPublisher'], hasKey('tools'))
        assertThat(publisherStepOptions['ESLintPublisher']['tools'], hasItem(hasEntry('pattern', '**/eslint.xml')))
        // ensure nothing else is published
        assertThat(publisherStepOptions, not(hasKey('PmdPublisher')))
        assertThat(publisherStepOptions, not(hasKey('DryPublisher')))
        assertThat(publisherStepOptions, not(hasKey('FindBugsPublisher')))
        assertThat(publisherStepOptions, not(hasKey('CheckStylePublisher')))
        assertThat(publisherStepOptions, not(hasKey('PyLintPublisher')))
        assertThat(publisherStepOptions, not(hasKey('TaskPublisher')))
    }

    @Test
    void testPublishForPythonWithDefaultSettings() throws Exception {
        // test
        stepRule.step.checksPublishResults(script: nullScript, pylint: true)
        // assert
        assertThat(publisherStepOptions, hasKey('PyLintPublisher'))
        // ensure correct parser is set set
        assertThat(publisherStepOptions['PyLintPublisher'], hasKey('tools'))
        assertThat(publisherStepOptions['PyLintPublisher']['tools'], hasItem(hasEntry('pattern', '**/pylint.log')))
        // ensure nothing else is published
        assertThat(publisherStepOptions, not(hasKey('PmdPublisher')))
        assertThat(publisherStepOptions, not(hasKey('DryPublisher')))
        assertThat(publisherStepOptions, not(hasKey('FindBugsPublisher')))
        assertThat(publisherStepOptions, not(hasKey('CheckStylePublisher')))
        assertThat(publisherStepOptions, not(hasKey('ESLintPublisher')))
        assertThat(publisherStepOptions, not(hasKey('TaskPublisher')))
    }

    @Test
    void testPublishNothingExplicitFalse() throws Exception {
        // test
        stepRule.step.checksPublishResults(script: nullScript, pmd: false)
        // assert
        // ensure pmd is not published
        assertThat(publisherStepOptions, not(hasKey('PmdPublisher')))
    }

    @Test
    void testPublishNothingImplicitTrue() throws Exception {
        // test
        stepRule.step.checksPublishResults(script: nullScript, pmd: [:])
        // assert
        // ensure pmd is published
        assertThat(publisherStepOptions, hasKey('PmdPublisher'))
    }

    @Test
    void testPublishNothingExplicitActiveFalse() throws Exception {
        // test
        stepRule.step.checksPublishResults(script: nullScript, pmd: [active: false])
        // assert
        // ensure pmd is not published
        assertThat(publisherStepOptions, not(hasKey('PmdPublisher')))
    }

    @Test
    void testPublishWithChangedStepDefaultSettings() throws Exception {
        // init
        // pmd has been set to active: true in step configuration
        nullScript.commonPipelineEnvironment.configuration =
        [
            steps: [
                checksPublishResults: [
                    pmd: [active: true]
                ]
            ]
        ]
        // test
        stepRule.step.checksPublishResults(script: nullScript)
        // assert
        assertThat(publisherStepOptions, hasKey('PmdPublisher'))
        // ensure nothing else is published
        assertThat(publisherStepOptions, not(hasKey('DryPublisher')))
        assertThat(publisherStepOptions, not(hasKey('FindBugsPublisher')))
        assertThat(publisherStepOptions, not(hasKey('CheckStylePublisher')))
        assertThat(publisherStepOptions, not(hasKey('ESLintPublisher')))
        assertThat(publisherStepOptions, not(hasKey('PyLintPublisher')))
        assertThat(publisherStepOptions, not(hasKey('TaskPublisher')))
    }

    @Test
    void testPublishWithCustomPattern() throws Exception {
        // test
        stepRule.step.checksPublishResults(script: nullScript, eslint: [pattern: 'my-fancy-file.ext'], pmd: [pattern: 'this-is-not-a-patter.xml'])
        // assert
        assertThat(publisherStepOptions, hasKey('PmdPublisher'))
        assertThat(publisherStepOptions, hasKey('ESLintPublisher'))
        // ensure custom patterns are set
        assertThat(publisherStepOptions['PmdPublisher'], hasKey('tools'))
        assertThat(publisherStepOptions['PmdPublisher']['tools'], hasItem(hasEntry('pattern', 'this-is-not-a-patter.xml')))
        assertThat(publisherStepOptions['ESLintPublisher'], hasKey('tools'))
        assertThat(publisherStepOptions['ESLintPublisher']['tools'], hasItem(hasEntry('pattern', 'my-fancy-file.ext')))
        // ensure nothing else is published
        assertThat(publisherStepOptions, not(hasKey('DryPublisher')))
        assertThat(publisherStepOptions, not(hasKey('FindBugsPublisher')))
        assertThat(publisherStepOptions, not(hasKey('CheckStylePublisher')))
        assertThat(publisherStepOptions, not(hasKey('PyLintPublisher')))
        assertThat(publisherStepOptions, not(hasKey('TaskPublisher')))
    }

    @Test
    void testPublishWithArchive() throws Exception {
        // test
        stepRule.step.checksPublishResults(script: nullScript, archive: true, eslint: true, pmd: true, cpd: true, findbugs: true, checkstyle: true)
        // assert
        assertThat(archiveStepPatterns, hasSize(5))
        assertThat(archiveStepPatterns, allOf(
            hasItem('**/target/pmd.xml'),
            hasItem('**/target/cpd.xml'),
            hasItem('**/target/findbugsXml.xml, **/target/findbugs.xml'),
            hasItem('**/target/checkstyle-result.xml'),
            hasItem('**/eslint.xml'),
        ))
    }

    @Test
    void testPublishWithPartialArchive() throws Exception {
        // test
        stepRule.step.checksPublishResults(script: nullScript, archive: true, eslint: [archive: false], pmd: true, cpd: true, findbugs: true, checkstyle: true)
        // assert
        assertThat(archiveStepPatterns, hasSize(4))
        assertThat(archiveStepPatterns, allOf(
            hasItem('**/target/pmd.xml'),
            hasItem('**/target/cpd.xml'),
            hasItem('**/target/findbugsXml.xml, **/target/findbugs.xml'),
            hasItem('**/target/checkstyle-result.xml'),
            not(hasItem('**/eslint.xml')),
        ))
    }

    @Test
    void testPublishWithDefaultThresholds() throws Exception {
        // test
        stepRule.step.checksPublishResults(script: nullScript, pmd: true)
        // assert
        assertThat(publisherStepOptions, hasKey('PmdPublisher'))
        assertThat(publisherStepOptions['PmdPublisher'], hasKey('qualityGates'))
        assertThat(publisherStepOptions['PmdPublisher']['qualityGates'], allOf(
            hasSize(2),
            hasItem(allOf(
                hasEntry('threshold', 1),
                hasEntry('type', 'TOTAL_ERROR'),
                hasEntry('unstable', false),
            )),
            hasItem(allOf(
                hasEntry('threshold', 1),
                hasEntry('type', 'TOTAL_HIGH'),
                hasEntry('unstable', false),
            )),
        ))
    }

    @Test
    void testPublishWithLegacyThresholds() throws Exception {
        // test
        stepRule.step.checksPublishResults(script: nullScript, pmd: [thresholds: [fail: [high: '10']]])
        // assert
        assertThat(publisherStepOptions, hasKey('PmdPublisher'))
        assertThat(publisherStepOptions['PmdPublisher'], hasKey('qualityGates'))
        assertThat(publisherStepOptions['PmdPublisher']['qualityGates'], allOf(
            //TODO: thresholds are added to existing qualityGates, thus we have 2 defined in the end
            hasSize(3),
            hasItem(allOf(
                hasEntry('threshold', 1),
                hasEntry('type', 'TOTAL_ERROR'),
                hasEntry('unstable', false),
            )),
            hasItem(allOf(
                hasEntry('threshold', 1),
                hasEntry('type', 'TOTAL_HIGH'),
                hasEntry('unstable', false),
            )),
            hasItem(allOf(
                hasEntry('threshold', 11),
                hasEntry('type', 'TOTAL_HIGH'),
                hasEntry('unstable', false),
            )),
        ))
    }

    @Test
    void testPublishWithCustomThresholds() throws Exception {
        // test
        stepRule.step.checksPublishResults(script: nullScript, pmd: [active: true, qualityGates: [[threshold: 20, type: 'TOTAL_LOW', unstable: false],[threshold: 10, type: 'TOTAL_NORMAL', unstable: false]]])
        // assert
        assertThat(publisherStepOptions, hasKey('PmdPublisher'))
        assertThat(publisherStepOptions['PmdPublisher'], hasKey('qualityGates'))
        assertThat(publisherStepOptions['PmdPublisher']['qualityGates'], allOf(
            hasSize(2),
            hasItem(allOf(
                hasEntry('threshold', 10),
                hasEntry('type', 'TOTAL_NORMAL'),
                hasEntry('unstable', false),
            )),
            hasItem(allOf(
                hasEntry('threshold', 20),
                hasEntry('type', 'TOTAL_LOW'),
                hasEntry('unstable', false),
            )),
        ))
    }
}
