package templates

import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsReadYamlRule
import util.JenkinsStepRule
import util.PipelineWhenException
import util.Rules

import static org.hamcrest.Matchers.*
import static org.junit.Assert.assertThat

class AbapEnvironmentPipelineStageBuildTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jsr)

    private stepsCalled = []

    @Before
    void init()  {
        binding.variables.env.STAGE_NAME = 'Build'

        helper.registerAllowedMethod('piperStageWrapper', [Map.class, Closure.class], {m, body ->
            assertThat(m.stageName, is('Build'))
            return body()
        })
        helper.registerAllowedMethod('cloudFoundryCreateServiceKey', [Map.class], {m -> stepsCalled.add('cloudFoundryCreateServiceKey')})
        helper.registerAllowedMethod('abapEnvironmentAssemblePackages', [Map.class], {m -> stepsCalled.add('abapEnvironmentAssemblePackages')})
        helper.registerAllowedMethod('abapEnvironmentBuild', [Map.class], {m -> stepsCalled.add('abapEnvironmentBuild')})
        helper.registerAllowedMethod('abapAddonAssemblyKitRegisterPackages', [Map.class], {m -> stepsCalled.add('abapAddonAssemblyKitRegisterPackages')})
        helper.registerAllowedMethod('abapAddonAssemblyKitReleasePackages', [Map.class], {m -> stepsCalled.add('abapAddonAssemblyKitReleasePackages')})
        helper.registerAllowedMethod('abapEnvironmentAssembleConfirm', [Map.class], {m -> stepsCalled.add('abapEnvironmentAssembleConfirm')})
        helper.registerAllowedMethod('abapAddonAssemblyKitCreateTargetVector', [Map.class], {m -> stepsCalled.add('abapAddonAssemblyKitCreateTargetVector')})
        helper.registerAllowedMethod('abapAddonAssemblyKitPublishTargetVector', [Map.class], {m -> stepsCalled.add('abapAddonAssemblyKitPublishTargetVector')})
        helper.registerAllowedMethod('abapEnvironmentCreateTag', [Map.class], {m -> stepsCalled.add('abapEnvironmentCreateTag')})
    }

    @Test
    void testAbapEnvironmentRunTestsWithoutHost() {
        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Build': true,
        ]
        jsr.step.abapEnvironmentPipelineStageBuild(script: nullScript)

        assertThat(stepsCalled, hasItems('cloudFoundryCreateServiceKey',
                                            'abapEnvironmentAssemblePackages',
                                            'abapEnvironmentBuild',
                                            'abapAddonAssemblyKitRegisterPackages',
                                            'abapAddonAssemblyKitReleasePackages',
                                            'abapEnvironmentAssembleConfirm',
                                            'abapAddonAssemblyKitCreateTargetVector',
                                            'abapAddonAssemblyKitPublishTargetVector'))
        assertThat(stepsCalled, not(hasItems('abapEnvironmentCreateTag')))
    }

    @Test
    void testAbapEnvironmentRunTestsWithHost() {
        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Build': true
        ]
        jsr.step.abapEnvironmentPipelineStageBuild(script: nullScript,  host: 'abc.com', generateTagForAddonComponentVersion: true)

        assertThat(stepsCalled, hasItems('abapEnvironmentAssemblePackages',
                                            'abapEnvironmentBuild',
                                            'abapAddonAssemblyKitRegisterPackages',
                                            'abapAddonAssemblyKitReleasePackages',
                                            'abapEnvironmentAssembleConfirm',
                                            'abapAddonAssemblyKitCreateTargetVector',
                                            'abapAddonAssemblyKitPublishTargetVector',
                                            'abapEnvironmentCreateTag'))
        assertThat(stepsCalled, not(hasItems('cloudFoundryCreateServiceKey')))
    }

    @Test
    void testAbapEnvironmentRunTest4TestBuild() {
        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Build': true
        ]
        jsr.step.abapEnvironmentPipelineStageBuild(script: nullScript, testBuild: true, generateTagForAddonComponentVersion: true)

        assertThat(stepsCalled, hasItems('cloudFoundryCreateServiceKey',
                                            'abapEnvironmentAssemblePackages',
                                            'abapEnvironmentBuild',
                                            'abapAddonAssemblyKitRegisterPackages',
                                            'abapAddonAssemblyKitCreateTargetVector'))
        assertThat(stepsCalled, not(hasItems('abapAddonAssemblyKitReleasePackages',
                                                'abapEnvironmentAssembleConfirm',
                                                'abapAddonAssemblyKitPublishTargetVector',
                                                'abapEnvironmentCreateTag')))
    }

}
