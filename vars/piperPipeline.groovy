void call(parameters) {
    pipeline {
        agent none
        triggers {
            issueCommentTrigger('.*/piper ([a-z]*).*')
        }
        options {
            skipDefaultCheckout()
            timestamps()
        }
        stages {
            stage('Init') {
                steps {
                    piperPipelineStageInit script: parameters.script, customDefaults: ['com.sap.piper/pipeline/stageOrdinals.yml'].plus(parameters.customDefaults ?: [])
                }
            }
            stage('Pull-Request Voting') {
                when { anyOf { branch 'PR-*'; branch parameters.script.commonPipelineEnvironment.getStepConfiguration('piperPipelineStagePRVoting', 'Pull-Request Voting').customVotingBranch } }
                steps {
                    piperPipelineStagePRVoting script: parameters.script
                }
            }
            stage('Build') {
                when {branch parameters.script.commonPipelineEnvironment.getStepConfiguration('', '').productiveBranch}
                steps {
                    piperPipelineStageBuild script: parameters.script
                }
            }
            stage('Additional Unit Tests') {
                when {allOf {branch parameters.script.commonPipelineEnvironment.getStepConfiguration('', '').productiveBranch; expression {return parameters.script.commonPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME)}}}
                steps {
                    piperPipelineStageAdditionalUnitTests script: parameters.script
                }
            }
            stage('Integration') {
                when {allOf {branch parameters.script.commonPipelineEnvironment.getStepConfiguration('', '').productiveBranch; expression {return parameters.script.commonPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME)}}}
                steps {
                    piperPipelineStageIntegration script: parameters.script
                }
            }
            stage('Acceptance') {
                when {allOf {branch parameters.script.commonPipelineEnvironment.getStepConfiguration('', '').productiveBranch; expression {return parameters.script.commonPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME)}}}
                steps {
                    piperPipelineStageAcceptance script: parameters.script
                }
            }
            stage('Security') {
                when {allOf {branch parameters.script.commonPipelineEnvironment.getStepConfiguration('', '').productiveBranch; expression {return parameters.script.commonPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME)}}}
                steps {
                    piperPipelineStageSecurity script: parameters.script
                }
            }
            stage('Performance') {
                when {allOf {branch parameters.script.commonPipelineEnvironment.getStepConfiguration('', '').productiveBranch; expression {return parameters.script.commonPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME)}}}
                steps {
                    piperPipelineStagePerformance script: parameters.script
                }
            }
            stage('Compliance') {
                when {allOf {branch parameters.script.commonPipelineEnvironment.getStepConfiguration('', '').productiveBranch; expression {return parameters.script.commonPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME)}}}
                steps {
                    piperPipelineStageCompliance script: parameters.script
                }
            }
            stage('Confirm') {
                agent none
                when {allOf {expression { env.BRANCH_NAME ==~ parameters.script.commonPipelineEnvironment.getStepConfiguration('', '').productiveBranch }; anyOf {expression {return (currentBuild.result == 'UNSTABLE')}; expression {return parameters.script.commonPipelineEnvironment.getStepConfiguration('piperInitRunStageConfiguration', env.STAGE_NAME).manualConfirmation}}}}
                steps {
                    piperPipelineStageConfirm script: parameters.script
                }
            }
            stage('Promote') {
                when { branch parameters.script.commonPipelineEnvironment.getStepConfiguration('', '').productiveBranch}
                steps {
                    piperPipelineStagePromote script: parameters.script
                }
            }
            stage('Release') {
                when {allOf {branch parameters.script.commonPipelineEnvironment.getStepConfiguration('', '').productiveBranch; expression {return parameters.script.commonPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME)}}}
                steps {
                    piperPipelineStageRelease script: parameters.script
                }
            }
        }
        post {
            /* https://jenkins.io/doc/book/pipeline/syntax/#post */
            success {buildSetResult(currentBuild)}
            aborted {buildSetResult(currentBuild, 'ABORTED')}
            failure {buildSetResult(currentBuild, 'FAILURE')}
            unstable {buildSetResult(currentBuild, 'UNSTABLE')}
            cleanup {
                piperPipelineStagePost script: parameters.script
            }
        }
    }
}
