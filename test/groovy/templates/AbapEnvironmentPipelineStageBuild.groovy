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

class AbapEnvironmentPipelineStageAUnitTest extends BasePiperTest {
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
        helper.registerAllowedMethod('abapAddonAssemblyKitReserveNextPackages', [Map.class], {m -> stepsCalled.add('abapAddonAssemblyKitReserveNextPackages')})
        helper.registerAllowedMethod('abapEnvironmentAssemblePackages', [Map.class], {m -> stepsCalled.add('abapEnvironmentAssemblePackages')})
        helper.registerAllowedMethod('abapAddonAssemblyKitRegisterPackages', [Map.class], {m -> stepsCalled.add('abapAddonAssemblyKitRegisterPackages')})
        helper.registerAllowedMethod('abapAddonAssemblyKitReleasePackages', [Map.class], {m -> stepsCalled.add('abapAddonAssemblyKitReleasePackages')})
        helper.registerAllowedMethod('abapEnvironmentAssembleConfirm', [Map.class], {m -> stepsCalled.add('abapEnvironmentAssembleConfirm')})
        helper.registerAllowedMethod('abapAddonAssemblyKitCreateTargetVector', [Map.class], {m -> stepsCalled.add('abapAddonAssemblyKitCreateTargetVector')})
        helper.registerAllowedMethod('abapAddonAssemblyKitPublishTargetVector', [Map.class], {m -> stepsCalled.add('abapAddonAssemblyKitPublishTargetVector')})
    }
    
    @Test
    void testAbapEnvironmentRunTestsWithoutHost() {
        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Build': true
        ]
        jsr.step.abapEnvironmentPipelineStageAUnit(script: nullScript)

        assertThat(stepsCalled, hasItems('cloudFoundryCreateServiceKey',
                                            'abapAddonAssemblyKitReserveNextPackages',
                                            'abapEnvironmentAssemblePackages',
                                            'abapAddonAssemblyKitRegisterPackages',
                                            'abapAddonAssemblyKitReleasePackages',
                                            'abapEnvironmentAssembleConfirm',
                                            'abapAddonAssemblyKitCreateTargetVector',
                                            'abapAddonAssemblyKitPublishTargetVector'))
    }

    @Test
    void testAbapEnvironmentRunTestsWithHost() {
        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Build': true
        ]
        jsr.step.abapEnvironmentPipelineStageAUnit(script: nullScript,  host: 'abc.com')

        assertThat(stepsCalled, hasItems('abapAddonAssemblyKitReserveNextPackages',
                                            'abapEnvironmentAssemblePackages',
                                            'abapAddonAssemblyKitRegisterPackages',
                                            'abapAddonAssemblyKitReleasePackages',
                                            'abapEnvironmentAssembleConfirm',
                                            'abapAddonAssemblyKitCreateTargetVector',
                                            'abapAddonAssemblyKitPublishTargetVector'))
        assertThat(stepsCalled, not(hasItems('cloudFoundryCreateServiceKey')))
    }

}
