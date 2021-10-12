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

class AbapEnvironmentPipelineTest extends BasePiperTest {
    private JenkinsStepRule jsr = new JenkinsStepRule(this)

    @Rule
    public RuleChain rules = Rules
        .getCommonRules(this)
        .around(new JenkinsReadYamlRule(this))
        .around(jsr)

    private skipDefaultCheckout = false
    private timestamps = false
    private stagesExecuted = []
    private stepsCalled = []

    @Before
    void init() {

        helper.registerAllowedMethod('library', [String.class], null)

        helper.registerAllowedMethod('pipeline', [Closure.class], null)

        helper.registerAllowedMethod('agent', [Closure.class], null)
        binding.setVariable('any', {})
        binding.setVariable('none', {})

        helper.registerAllowedMethod('options', [Closure.class], null)
        helper.registerAllowedMethod('skipDefaultCheckout', [], {skipDefaultCheckout = true})

        helper.registerAllowedMethod('stages', [Closure.class], null)

        helper.registerAllowedMethod('stage', [String.class, Closure.class], {stageName, body ->

            def stageResult

            binding.variables.env.STAGE_NAME = stageName

            helper.registerAllowedMethod('when', [Closure.class], {cWhen ->

                helper.registerAllowedMethod('allOf', [Closure.class], {cAllOf ->
                    def branchResult = false
                    helper.registerAllowedMethod('branch', [String.class], {branchName  ->
                        if (!branchResult)
                            branchResult = (branchName == env.BRANCH_NAME)
                        if( !branchResult) {
                            throw new PipelineWhenException("Stage '${stageName}' skipped - expression: '${branchResult}'")
                        }
                    })
                    helper.registerAllowedMethod('expression', [Closure.class], { Closure cExp ->
                        def result = cExp()
                        if(!result) {
                            throw new PipelineWhenException("Stage '${stageName}' skipped - expression: '${result}'")
                        }
                        return result
                    })
                    return cAllOf()
                })

                helper.registerAllowedMethod('anyOf', [Closure.class], {cAnyOf ->
                    def result = false
                    helper.registerAllowedMethod('branch', [String.class], {branchName  ->
                        if (!result)
                            result = (branchName == env.BRANCH_NAME)
                        return result
                    })
                    helper.registerAllowedMethod('expression', [Closure.class], { Closure cExp ->
                        if (!result)
                            result = cExp()
                        return result
                    })
                    cAnyOf()
                    if(!result) {
                        throw new PipelineWhenException("Stage '${stageName}' skipped - anyOf: '${result}'")
                    }
                    return cAnyOf()
                })

                helper.registerAllowedMethod('branch', [String.class], {branchName  ->
                    def result =  (branchName == env.BRANCH_NAME)
                    if(result == false) {
                        throw new PipelineWhenException("Stage '${stageName}' skipped - expected branch: '${branchName}' while current branch: '${env.BRANCH_NAME}'")
                    }
                    return result
                })

                helper.registerAllowedMethod('expression', [Closure.class], { Closure cExp ->
                    def result = cExp()
                    if(!result) {
                        throw new PipelineWhenException("Stage '${stageName}' skipped - expression: '${result}'")
                    }
                    return result
                })
                return cWhen()
            })

            // Stage is not executed if build fails or aborts
            def status = currentBuild.result
            switch (status) {
                case 'FAILURE':
                case 'ABORTED':
                    break
                default:
                    try {
                        stageResult = body()
                        stagesExecuted.add(stageName)
                    }
                    catch (PipelineWhenException pwe) {
                        //skip stage due to not met when expression
                    }
                    catch (Exception e) {
                        throw e
                    }
            }
            return stageResult
        })

        helper.registerAllowedMethod('steps', [Closure], null)
        helper.registerAllowedMethod('parallel', [Closure], null)
        helper.registerAllowedMethod('post', [Closure], {c -> c()})
        helper.registerAllowedMethod('success', [Closure], {c -> c()})
        helper.registerAllowedMethod('failure', [Closure], {c -> c()})
        helper.registerAllowedMethod('aborted', [Closure], {c -> c()})
        helper.registerAllowedMethod('unstable', [Closure], {c -> c()})
        helper.registerAllowedMethod('unsuccessful', [Closure], {c -> c()})
        helper.registerAllowedMethod('cleanup', [Closure], {c -> c()})
        helper.registerAllowedMethod('input', [Map], {m -> return null})

        helper.registerAllowedMethod('abapEnvironmentPipelineStageInit', [Map.class], {m ->
            stepsCalled.add('abapEnvironmentPipelineStageInit')
        })

        helper.registerAllowedMethod('abapEnvironmentPipelineStagePrepareSystem', [Map.class], {m ->
            stepsCalled.add('abapEnvironmentPipelineStagePrepareSystem')
        })

        helper.registerAllowedMethod('abapEnvironmentPipelineStageCloneRepositories', [Map.class], {m ->
            stepsCalled.add('abapEnvironmentPipelineStageCloneRepositories')
        })

        helper.registerAllowedMethod('abapEnvironmentPipelineStageAUnit', [Map.class], {m ->
            stepsCalled.add('abapEnvironmentPipelineStageAUnit')
        })

        helper.registerAllowedMethod('abapEnvironmentPipelineStageATC', [Map.class], {m ->
            stepsCalled.add('abapEnvironmentPipelineStageATC')
        })

        helper.registerAllowedMethod('abapEnvironmentPipelineStagePost', [Map.class], {m ->
            stepsCalled.add('abapEnvironmentPipelineStagePost')
        })

        helper.registerAllowedMethod('abapEnvironmentPipelineStageBuild', [Map.class], {m ->
            stepsCalled.add('abapEnvironmentPipelineStageBuild')
        })

        helper.registerAllowedMethod('abapEnvironmentPipelineStageInitialChecks', [Map.class], {m ->
            stepsCalled.add('abapEnvironmentPipelineStageInitialChecks')
        })

        helper.registerAllowedMethod('abapEnvironmentPipelineStagePublish', [Map.class], {m ->
            stepsCalled.add('abapEnvironmentPipelineStagePublish')
        })

        helper.registerAllowedMethod('abapEnvironmentPipelineStageConfirm', [Map.class], {m ->
            stepsCalled.add('abapEnvironmentPipelineStageConfirm')
        })

        helper.registerAllowedMethod('abapEnvironmentPipelineStageIntegrationTests', [Map.class], {m ->
            stepsCalled.add('abapEnvironmentPipelineStageIntegrationTests')
        })

        nullScript.prepareDefaultValues(script: nullScript)

    }

    @Test
    void testAbapEnvironmentPipelineNoPrepareNoCleanup() {


        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Clone Repositories': true,
            'Tests': true,
            'ATC': true,
            'AUnit': true,
        ]

        jsr.step.abapEnvironmentPipeline(script: nullScript)

        assertThat(stepsCalled, hasItems(
            'abapEnvironmentPipelineStageInit',
            'abapEnvironmentPipelineStageCloneRepositories',
            'abapEnvironmentPipelineStageATC',
            'abapEnvironmentPipelineStageAUnit',
            'abapEnvironmentPipelineStagePost'
        ))
        assertThat(stepsCalled, not(hasItem('abapEnvironmentPipelineStagePrepareSystem')))
    }

    @Test
    void testAbapEnvironmentPipelineNoCloneNoRunTests() {

        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Prepare System': true,
        ]
        jsr.step.abapEnvironmentPipeline(script: nullScript)

        assertThat(stepsCalled, hasItems(
            'abapEnvironmentPipelineStageInit',
            'abapEnvironmentPipelineStagePrepareSystem',
            'abapEnvironmentPipelineStagePost'
        ))
        assertThat(stepsCalled, not(hasItem('abapEnvironmentPipelineStageCloneRepositories')))
        assertThat(stepsCalled, not(hasItem('abapEnvironmentPipelineStageATC')))
        assertThat(stepsCalled, not(hasItem('abapEnvironmentPipelineStageAUnit')))
    }

    @Test
    void testAbapEnvironmentPipelineAllOn() {

        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'Prepare System': true,
            'Clone Repositories': true,
            'Tests': true,
            'ATC': true,
            'AUnit': true,
            'Build': true,
            'Integration Tests': true,
            'Publish': true
        ]
        jsr.step.abapEnvironmentPipeline(script: nullScript)

        assertThat(stepsCalled, hasItems(
            'abapEnvironmentPipelineStageInit',
            'abapEnvironmentPipelineStagePrepareSystem',
            'abapEnvironmentPipelineStageCloneRepositories',
            'abapEnvironmentPipelineStageATC',
            'abapEnvironmentPipelineStageAUnit',
            'abapEnvironmentPipelineStagePost',
            'abapEnvironmentPipelineStageBuild',
            'abapEnvironmentPipelineStageInitialChecks',
            'abapEnvironmentPipelineStagePublish',
            'abapEnvironmentPipelineStageConfirm',
            'abapEnvironmentPipelineStageIntegrationTests'
        ))
    }

    @Test
    void testAbapEnvironmentPipelineATCOnly() {

        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'ATC': true,
        ]
        jsr.step.abapEnvironmentPipeline(script: nullScript)

        assertThat(stepsCalled, hasItems(
            'abapEnvironmentPipelineStageInit',
            'abapEnvironmentPipelineStageATC',
            'abapEnvironmentPipelineStagePost'
        ))
        assertThat(stepsCalled, not(hasItems(
            'abapEnvironmentPipelineStageAUnit'
        )))
    }
    @Test
    void testAbapEnvironmentPipelineAunitOnly() {

        nullScript.commonPipelineEnvironment.configuration.runStage = [
            'AUnit': true,
        ]
        jsr.step.abapEnvironmentPipeline(script: nullScript)

        assertThat(stepsCalled, hasItems(
            'abapEnvironmentPipelineStageInit',
            'abapEnvironmentPipelineStageAUnit',
            'abapEnvironmentPipelineStagePost'
        ))
        assertThat(stepsCalled, not(hasItems(
            'abapEnvironmentPipelineStageATC'
        )))
    }
}
