void call(parameters) {
    pipeline {
        agent none
        options {
            skipDefaultCheckout()
        }
        stages {

            stage('Init') {
                steps {
                    abapEnvironmentPipelineStageInit script: parameters.script, customDefaults: ['com.sap.piper/pipeline/abapEnvironmentPipelineDefaults.yml'].plus(parameters.customDefaults ?: [])
                }
            }

            stage('Initial Checks') {
                when {expression {return parameters.script.commonPipelineEnvironment.configuration.runStage?.get("Build")}}
                steps {
                    abapEnvironmentPipelineStageInitialChecks script: parameters.script
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

            stage('Test') {
                    parallel {
                        stage('ATC') {
                            when {expression {return parameters.script.commonPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME)}}
                            steps {
                                abapEnvironmentPipelineStageATC script: parameters.script
                            }
                        }
                        stage('AUnit') {
                            when {expression {return parameters.script.commonPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME)}}
                            steps {
                                abapEnvironmentPipelineStageAUnit script: parameters.script
                            }
                        }
                    }
            }

            stage('Build') {
                when {expression {return parameters.script.commonPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME)}}
                steps {
                    abapEnvironmentPipelineStageBuild script: parameters.script
                }
            }

            stage('Integration Tests') {
                when {expression {return parameters.script.commonPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME)}}
                steps {
                    abapEnvironmentPipelineStageIntegrationTests script: parameters.script
                }
            }

            stage('Confirm') {
                when {expression {return parameters.script.commonPipelineEnvironment.configuration.runStage?.get("Publish")}}
                steps {
                    abapEnvironmentPipelineStageConfirm script: parameters.script
                }
            }

            stage('Publish') {
                when {expression {return parameters.script.commonPipelineEnvironment.configuration.runStage?.get(env.STAGE_NAME)}}
                steps {
                    abapEnvironmentPipelineStagePublish script: parameters.script
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
                abapEnvironmentPipelineStagePost script: parameters.script
            }
        }
    }
}
