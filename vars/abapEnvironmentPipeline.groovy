void call(parameters) {
    pipeline {
        agent any
        stages {

            stage('Init') {
                steps {
                    abapEnvironmentPipelineStageInit script: parameters.script
                }
            }

            stage('Prepare System') {
                when {expression {return parameters.script.commonPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME)}}
                steps {
                    abapEnvironmentPipelineStagePrepareSystem script: parameters.script
                }
            }

            stage('Clone Repositories') {
                when {expression {return parameters.script.commonPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME)}}
                steps {
                    abapEnvironmentPipelineStageCloneRepositories script: parameters.script
                }
            }

            stage('ATC') {
                when {expression {return parameters.script.commonPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME)}}
                steps {
                    abapEnvironmentPipelineStageATC script: parameters.script
                }
            }
        }
        post {
            /* https://jenkins.io/doc/book/pipeline/syntax/#post */
            success {buildSetResult(currentBuild)}
            aborted {buildSetResult(currentBuild, 'ABORTED')}
            failure {buildSetResult(currentBuild, 'FAILURE')}
            unstable {buildSetResult(currentBuild, 'UNSTABLE')}
            unsuccessful {
                input "Unsuccessful build: delete system?"
            }
            cleanup {
                abapEnvironmentPipelineStageCleanup script: parameters.script
            }
        }
    }
}
