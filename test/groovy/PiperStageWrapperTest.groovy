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

    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)

    private Map lockMap = [:]
    private int countNodeUsage = 0
    private String nodeLabel = ''

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(loggingRule)
        .around(stepRule)

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
    void testDefault() {
        def executed = false
        stepRule.step.piperStageWrapper(
            script: nullScript,
            juStabUtils: utils,
            ordinal: 10,
            stageName: 'test'

        ) {
            executed = true
        }
        assertThat(executed, is(true))
        assertThat(lockMap.size(), is(2))
        assertThat(countNodeUsage, is(1))
    }

    @Test
    void testNoLocking() {
        def executed = false
        stepRule.step.piperStageWrapper(
            script: nullScript,
            juStabUtils: utils,
            nodeLabel: 'testLabel',
            ordinal: 10,
            stageLocking: false,
            stageName: 'test'

        ) {
            executed = true
        }
        assertThat(executed, is(true))
        assertThat(lockMap.size(), is(0))
        assertThat(countNodeUsage, is(1))
        assertThat(nodeLabel, is('testLabel'))
    }

    @Test
    void testStageExit() {
        helper.registerAllowedMethod('fileExists', [String.class], {s ->
            return (s == '.pipeline/extensions/test.groovy')
        })

        helper.registerAllowedMethod('load', [String.class], {
            return helper.loadScript('test/resources/stages/test.groovy')
        })
        nullScript.commonPipelineEnvironment.gitBranch = 'testBranch'

        def executed = false
        stepRule.step.piperStageWrapper(
            script: nullScript,
            juStabUtils: utils,
            ordinal: 10,
            stageName: 'test'
        ) {
            executed = true
        }

        assertThat(executed, is(true))
        assertThat(loggingRule.log, containsString('[piperStageWrapper] Running project interceptor \'.pipeline/extensions/test.groovy\' for test.'))
        assertThat(loggingRule.log, containsString('Stage Name: test'))
        assertThat(loggingRule.log, containsString('Config: ['))
        assertThat(loggingRule.log, containsString('testBranch'))
    }
}

