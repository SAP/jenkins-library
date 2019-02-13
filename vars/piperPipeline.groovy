void call(parameters) {
    pipeline {
        agent none
        options {
            skipDefaultCheckout()
            timestamps()
        }
        stages {
            stage('Init') {
                steps {
                    library 'piper-lib-os'
                    piperPipelineStageInit script: parameters.script, customDefaults: parameters.customDefaults
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
                when {allOf {branch parameters.script.commonPipelineEnvironment.getStepConfiguration('', '').productiveBranch; expression {return parameters.script.commonPipelineEnvironment.getStepConfiguration('piperInitRunStageConfiguration', env.STAGE_NAME).manualConfirmation}}}
                steps {
                    input message: 'Shall we proceed to promotion & release?'
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
            always {
                influxWriteData script: parameters.script, wrapInNode: true
                mailSendNotification script: parameters.script, wrapInNode: true
            }
        }
    }
}
