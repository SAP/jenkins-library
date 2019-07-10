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

class PiperPipelineTest extends BasePiperTest {
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
        helper.registerAllowedMethod('timestamps', [], {timestamps = true})

        helper.registerAllowedMethod('triggers', [Closure.class], null)
        helper.registerAllowedMethod('issueCommentTrigger', [String.class], { s ->
            assertThat(s, is('.*/piper ([a-z]*).*'))
        })

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
        helper.registerAllowedMethod('post', [Closure], {c -> c()})
        helper.registerAllowedMethod('success', [Closure], {c -> c()})
        helper.registerAllowedMethod('failure', [Closure], {c -> c()})
        helper.registerAllowedMethod('aborted', [Closure], {c -> c()})
        helper.registerAllowedMethod('unstable', [Closure], {c -> c()})
        helper.registerAllowedMethod('cleanup', [Closure], {c -> c()})

        helper.registerAllowedMethod('input', [Map], {m -> return null})

        helper.registerAllowedMethod('influxWriteData', [Map], {m ->
            assertThat(m.wrapInNode, is(true))
        })
        helper.registerAllowedMethod('mailSendNotification', [Map], {m ->
            assertThat(m.wrapInNode, is(true))
        })

        helper.registerAllowedMethod('piperPipelineStageInit', [Map.class], {m ->
            stepsCalled.add('piperPipelineStageInit')
        })
        helper.registerAllowedMethod('piperPipelineStagePRVoting', [Map.class], {m ->
            stepsCalled.add('piperPipelineStagePRVoting')
        })
        helper.registerAllowedMethod('piperPipelineStageBuild', [Map.class], {m ->
            stepsCalled.add('piperPipelineStageBuild')
        })
        helper.registerAllowedMethod('piperPipelineStageAdditionalUnitTests', [Map.class], {m ->
            stepsCalled.add('piperPipelineStageAdditionalUnitTests')
        })
        helper.registerAllowedMethod('piperPipelineStageIntegration', [Map.class], {m ->
            stepsCalled.add('piperPipelineStageIntegration')
        })
        helper.registerAllowedMethod('piperPipelineStageAcceptance', [Map.class], {m ->
            stepsCalled.add('piperPipelineStageAcceptance')
        })
        helper.registerAllowedMethod('piperPipelineStageSecurity', [Map.class], {m ->
            stepsCalled.add('piperPipelineStageSecurity')
        })
        helper.registerAllowedMethod('piperPipelineStagePerformance', [Map.class], {m ->
            stepsCalled.add('piperPipelineStagePerformance')
        })
        helper.registerAllowedMethod('piperPipelineStageCompliance', [Map.class], {m ->
            stepsCalled.add('piperPipelineStageCompliance')
        })
        helper.registerAllowedMethod('piperPipelineStageConfirm', [Map.class], {m ->
            stepsCalled.add('piperPipelineStageConfirm')
        })
        helper.registerAllowedMethod('piperPipelineStagePromote', [Map.class], {m ->
            stepsCalled.add('piperPipelineStagePromote')
        })
        helper.registerAllowedMethod('piperPipelineStageRelease', [Map.class], {m ->
            stepsCalled.add('piperPipelineStageRelease')
        })
        helper.registerAllowedMethod('piperPipelineStagePost', [Map.class], {m ->
            stepsCalled.add('piperPipelineStagePost')
        })

        nullScript.prepareDefaultValues(script: nullScript)

    }

    @Test
    void testPRVoting() {

        helper.registerAllowedMethod('piperPipelineStageInit', [Map], null)

        binding.variables.env.BRANCH_NAME = 'PR-*'

        nullScript.commonPipelineEnvironment.configuration = [runStage:[Integration:[test: 'test']]]
        jsr.step.piperPipeline(script: nullScript)

        assertThat(skipDefaultCheckout, is(true))
        assertThat(timestamps, is(true))

        assertThat(stagesExecuted.size(), is(2))
        assertThat(stagesExecuted, allOf(hasItem('Init'), hasItem('Pull-Request Voting')))

        assertThat(stepsCalled, hasItem('piperPipelineStagePRVoting'))
    }

    @Test
    void testConfirmUnstable() {
        nullScript.commonPipelineEnvironment.configuration = [
            general: [
                manualConfirmation: false
            ]
        ]
        binding.setVariable('currentBuild', [result: 'UNSTABLE'])

        jsr.step.piperPipeline(script: nullScript)

        assertThat(stepsCalled, hasItem('piperPipelineStageConfirm'))

    }

    @Test
    void testNoConfirm() {
        nullScript.commonPipelineEnvironment.configuration = [
            general: [
                manualConfirmation: false
            ]
        ]
        jsr.step.piperPipeline(script: nullScript)

        assertThat(stepsCalled, not(hasItem('piperPipelineStageConfirm')))
    }

    @Test
    void testMasterPipelineAllOn() {

        nullScript.commonPipelineEnvironment.configuration.runStage = [
            Build: true,
            'Additional Unit Tests': true,
            Integration: true,
            Acceptance: true,
            Security: true,
            Performance: true,
            Compliance: true,
            Promote: true,
            Release: true
        ]
        jsr.step.piperPipeline(script: nullScript)

        assertThat(stepsCalled, hasItems(
            'piperPipelineStageInit',
            'piperPipelineStageBuild',
            'piperPipelineStageAdditionalUnitTests',
            'piperPipelineStageIntegration',
            'piperPipelineStageAcceptance',
            'piperPipelineStageSecurity',
            'piperPipelineStagePerformance',
            'piperPipelineStageCompliance',
            'piperPipelineStageConfirm',
            'piperPipelineStagePromote',
            'piperPipelineStageRelease',
            'piperPipelineStagePost'
        ))
    }
}
