#!groovy
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

import static org.hamcrest.CoreMatchers.containsString
import static org.hamcrest.CoreMatchers.is
import static org.junit.Assert.assertThat

class PiperStageWrapperTest extends BasePiperTest {

    private JenkinsLoggingRule jlr = new JenkinsLoggingRule(this)
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    private Map lockMap = [:]
    private int countNodeUsage = 0
    private String nodeLabel = ''

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jlr)
        .around(jsr)

    @Before
    void init() throws Exception {
        helper.registerAllowedMethod('deleteDir', [], {return null})
        helper.registerAllowedMethod('lock', [Map.class, Closure.class], {m, body ->
            assertThat(m.resource.toString(), containsString('/10'))
            lockMap = m
            body()
        })
        helper.registerAllowedMethod('milestone', [Integer.class], {ordinal ->
            assertThat(ordinal, is(10))
        })
        helper.registerAllowedMethod('node', [String.class, Closure.class], {s, body ->
            nodeLabel = s
            countNodeUsage++
            body()

        })
        helper.registerAllowedMethod('fileExists', [String.class], {s ->
            return false
        })
    }

    @Test
    void testStageExitFilePath() {
        def config = [extensionLocation: '.pipeline/extensions/']
        assertThat(jsr.step.piperStageWrapper.stageExitFilePath('test Stage', config), is('.pipeline/extensions/test_stage.groovy'))
    }

    @Test
    void testDefault() {
        def testInt = 1
        jsr.step.piperStageWrapper(
            script: nullScript,
            juStabUtils: utils,
            ordinal: 10,
            stageName: 'test'

        ) {
            testInt ++
        }
        assertThat(testInt, is(2))
        assertThat(lockMap.size(), is(2))
        assertThat(countNodeUsage, is(1))
    }

    @Test
    void testNoLocking() {
        def testInt = 1
        jsr.step.piperStageWrapper(
            script: nullScript,
            juStabUtils: utils,
            nodeLabel: 'testLabel',
            ordinal: 10,
            stageLocking: false,
            stageName: 'test'

        ) {
            testInt ++
        }
        assertThat(testInt, is(2))
        assertThat(lockMap.size(), is(0))
        assertThat(countNodeUsage, is(1))
        assertThat(nodeLabel, is('testLabel'))
    }

    @Test
    void testNoNode() {
        def testInt = 1
        jsr.step.piperStageWrapper(
            script: nullScript,
            juStabUtils: utils,
            ordinal: 10,
            stageName: 'test',
            withNode: false
        ) {
            testInt ++
        }
        assertThat(testInt, is(2))
        assertThat(lockMap.size(), is(2))
        assertThat(countNodeUsage, is(0))
    }

    @Test
    void testStageExit() {
        helper.registerAllowedMethod('fileExists', [String.class], {s ->
            return true
        })

        helper.registerAllowedMethod('load', [String.class], {
            return helper.loadScript('test/resources/stages/test.groovy')
        })

        def testInt = 1
        jsr.step.piperStageWrapper(
            script: nullScript,
            juStabUtils: utils,
            ordinal: 10,
            stageName: 'test'
        ) {
            testInt ++
        }

        assertThat(testInt, is(2))
        assertThat(jlr.log, containsString('[piperStageWrapper] Running interceptor \'.pipeline/extensions/test.groovy\' for test.'))
        assertThat(jlr.log, containsString('Stage Name: test'))
        assertThat(jlr.log, containsString('Config 1:'))
        assertThat(jlr.log, containsString('Config 2:'))
    }
}

