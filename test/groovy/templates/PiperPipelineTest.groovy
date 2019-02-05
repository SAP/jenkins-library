#!groovy
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

        helper.registerAllowedMethod('stages', [Closure.class], null)

        helper.registerAllowedMethod('stage', [String.class, Closure.class], {stageName, body ->

            def stageResult

            binding.variables.env.STAGE_NAME = stageName

            helper.registerAllowedMethod('when', [Closure.class], {cWhen ->

                helper.registerAllowedMethod('allOf', [Closure.class], null)

                helper.registerAllowedMethod('anyOf', [Closure.class], {cAnyOf ->
                    def result = false
                    helper.registerAllowedMethod('branch', [String.class], {branchName  ->
                        if (!result)
                            result = (branchName == env.BRANCH_NAME)
                        if( !result) {
                            throw new PipelineWhenException("Stage '${stageName}' skipped - expression: '${result}'")
                        }
                    })
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
        helper.registerAllowedMethod('post', [Closure], null)
        helper.registerAllowedMethod('always', [Closure], null)

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
        helper.registerAllowedMethod('input', [Map.class], {m ->
            stepsCalled.add('input')
        })
        helper.registerAllowedMethod('piperPipelineStagePromote', [Map.class], {m ->
            stepsCalled.add('piperPipelineStagePromote')
        })
        helper.registerAllowedMethod('piperPipelineStageRelease', [Map.class], {m ->
            stepsCalled.add('piperPipelineStageRelease')
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
    void testConfirm() {
        jsr.step.piperPipeline(script: nullScript)

        assertThat(stepsCalled, hasItem('input'))

    }

    @Test
    void testNoConfirm() {
        nullScript.commonPipelineEnvironment.configuration = [
            general: [
                manualConfirmation: false
            ]
        ]
        jsr.step.piperPipeline(script: nullScript)

        assertThat(stepsCalled, not(hasItem('input')))
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
            'input',
            'piperPipelineStagePromote',
            'piperPipelineStageRelease'
        ))
    }
}
