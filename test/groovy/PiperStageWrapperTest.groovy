import com.sap.piper.DebugReport
import hudson.AbortException
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsEnvironmentRule
import util.JenkinsLoggingRule
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.Rules

import static org.hamcrest.CoreMatchers.containsString
import static org.hamcrest.Matchers.contains
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.not
import static org.junit.Assert.assertThat

class PiperStageWrapperTest extends BasePiperTest {

    private ExpectedException thrown = ExpectedException.none()

    private JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsStepRule stepRule = new JenkinsStepRule(this)

    private Map lockMap = [:]
    private int countNodeUsage = 0
    private String nodeLabel = ''
    private boolean executedOnKubernetes = false
    private List customEnv = []

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(thrown)
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

        helper.registerAllowedMethod('dockerExecuteOnKubernetes', [Map.class, Closure.class], {params, body ->
            executedOnKubernetes = true
            body()
        })

        helper.registerAllowedMethod('withEnv', [List.class, Closure.class], {env, body ->
            customEnv = env
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
        assertThat(executedOnKubernetes, is(false))
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
    void testExecuteStageOnKubernetes() {
        def executed = false

        binding.variables.env.ON_K8S = true
        nullScript.commonPipelineEnvironment.configuration = [general: [runStageInPod: true]]

        stepRule.step.piperStageWrapper(
            script: nullScript,
            juStabUtils: utils,
            stageName: 'test',
            ordinal: 10
        ) {
            executed = true
        }
        assertThat(executed, is(true))
        assertThat(executedOnKubernetes, is(true))
        assertThat(customEnv[1].toString(), is("POD_NAME=test"))
    }

    @Test
    void testStageNameInEnv() {
        def executed = false

        binding.variables.env.STAGE_NAME = 'label'

        stepRule.step.piperStageWrapper(
            script: nullScript,
            juStabUtils: utils,
            stageName: 'test',
            ordinal: 10
        ) {
            executed = true
        }
        assertThat(executed, is(true))
        assertThat(customEnv[0].toString(), is("STAGE_NAME=test"))
    }

    @Test
    void testStageNameAlreadyInEnv() {
        def executed = false

        binding.variables.env.STAGE_NAME = 'test'

        stepRule.step.piperStageWrapper(
            script: nullScript,
            juStabUtils: utils,
            ordinal: 10
        ) {
            executed = true
        }
        assertThat(executed, is(true))
        assertThat(customEnv.size(), is(0))
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

    @Test
    void testGlobalOverwritingExtension() {
        helper.registerAllowedMethod('fileExists', [String.class], {s ->
            return (s == '.pipeline/tmp/global_extensions/test_global_overwriting.groovy')
        })

        helper.registerAllowedMethod('load', [String.class], {
            return helper.loadScript('test/resources/stages/test_global_overwriting.groovy')
        })
        nullScript.commonPipelineEnvironment.gitBranch = 'testBranch'

        def executed = false
        stepRule.step.piperStageWrapper(
            script: nullScript,
            juStabUtils: utils,
            ordinal: 10,
            stageName: 'test_global_overwriting'
        ) {
            executed = true
        }

        assertThat(executed, is(false))
        assertThat(loggingRule.log, containsString('Stage Name: test_global_overwriting'))
        assertThat(loggingRule.log, containsString('Config: ['))
        assertThat(loggingRule.log, containsString('testBranch'))
        assertThat(loggingRule.log, containsString('Not calling test_global_overwriting'))
        assertThat(DebugReport.instance.globalExtensions.test_global_overwriting, is('Overwrites'))
    }

    @Test
    void testStageOldInterceptor() {
        helper.registerAllowedMethod('fileExists', [String.class], { path ->
            return (path == '.pipeline/extensions/test_old_extension.groovy')
        })

        helper.registerAllowedMethod('load', [String.class], {
            return helper.loadScript('test/resources/stages/test_old_extension.groovy')
        })
        nullScript.commonPipelineEnvironment.gitBranch = 'testBranch'

        def executed = false
        stepRule.step.piperStageWrapper(
            script: nullScript,
            juStabUtils: utils,
            ordinal: 10,
            stageName: 'test_old_extension'
        ) {
            executed = true
        }

        assertThat(executed, is(true))
        assertThat(loggingRule.log, containsString('[piperStageWrapper] Running project interceptor \'.pipeline/extensions/test_old_extension.groovy\' for test_old_extension.'))
        assertThat(loggingRule.log, containsString('[Warning] The interface to implement extensions has changed.'))
        assertThat(loggingRule.log, containsString('Stage Name: test_old_extension'))
        assertThat(loggingRule.log, containsString('Config: ['))
        assertThat(loggingRule.log, containsString('testBranch'))
        assertThat(DebugReport.instance.localExtensions.test_old_extension, is('Extends'))
    }

    @Test
    void testExtensionDeactivation() {
        helper.registerAllowedMethod('fileExists', [String.class], { path ->
            return (path == '.pipeline/extensions/test_old_extension.groovy')
        })
        helper.registerAllowedMethod('load', [String.class], {
            return helper.loadScript('test/resources/stages/test_old_extension.groovy')
        })

        nullScript.commonPipelineEnvironment.gitBranch = 'testBranch'
        nullScript.env = [PIPER_DISABLE_EXTENSIONS: 'true']
        stepRule.step.piperStageWrapper(
            script: nullScript,
            juStabUtils: utils,
            ordinal: 10,
            stageName: 'test_old_extension'
        ) {}
        //setting above parameter to 'true' bypasses the below message
        assertThat(loggingRule.log, not(containsString("[piperStageWrapper] Running project interceptor '.pipeline/extensions/test_old_extension.groovy' for test_old_extension.")))
    }

    @Test
    void testPipelineResilienceMandatoryStep() {
        thrown.expectMessage('expected error')

        nullScript.commonPipelineEnvironment.configuration = [general: [failOnError: false]]

        stepRule.step.piperStageWrapper (script: nullScript, stageLocking: false, stageName: 'testStage', juStabUtils: utils) {
            throw new AbortException('expected error')
        }
    }

    @Test
    void testStageCrashesInExtension() {
        helper.registerAllowedMethod('fileExists', [String.class], { path ->
            return (path == '.pipeline/tmp/global_extensions/test_crashing_extension.groovy')
        })

        helper.registerAllowedMethod('load', [String.class], {
            return helper.loadScript('test/resources/stages/test_crashing_extension.groovy')
        })

        Throwable caught = null
        def executed = false
        // Clear DebugReport to avoid left-overs from another UnitTest
        DebugReport.instance.failedBuild = [:]

        try {
            stepRule.step.piperStageWrapper(
                script: nullScript,
                juStabUtils: utils,
                ordinal: 10,
                stageName: 'test_crashing_extension'
            ) {
                executed = true
            }
        } catch (Throwable t) {
            caught = t
        }

        assertThat(executed, is(true))
        assertThat(loggingRule.log, containsString('[piperStageWrapper] Found global interceptor \'.pipeline/tmp/global_extensions/test_crashing_extension.groovy\' for test_crashing_extension.'))
        assertThat(DebugReport.instance.failedBuild.step, is('test_crashing_extension(extended)'))
        assertThat(DebugReport.instance.failedBuild.fatal, is('true'))
        assertThat(DebugReport.instance.failedBuild.reason, is(caught))
    }
}
