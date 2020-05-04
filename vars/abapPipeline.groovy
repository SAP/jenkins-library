void call(parameters) {
    pipeline {
        agent any
        //triggers {
        //    issueCommentTrigger('.*/piper ([a-z]*).*')
        //}
        options {
            timestamps()
        }
        stages {

            stage('Init') {
                steps {
            //     library 'piper-lib-os-dev'
            //     //setupCommonPipelineEnvironment script: parameters.script
            //     piperPipelineStageInit script: parameters.script, customDefaults: ['com.sap.piper/pipeline/stageOrdinals.yml'].plus(parameters.customDefaults ?: [])

                    abapPipelineInit script: parameters.script
                }
            }

            stage('Create Service') {
                steps {

                    cloudFoundryCreateService script: parameters.script
                }
            }
            stage('Delete Service') {
                steps {
                    cloudFoundryDeleteService script: parameters.script
                }
            }
        }
        // post {
        //     /* https://jenkins.io/doc/book/pipeline/syntax/#post */
        //     success {buildSetResult(currentBuild)}
        //     aborted {buildSetResult(currentBuild, 'ABORTED')}
        //     failure {buildSetResult(currentBuild, 'FAILURE')}
        //     unstable {buildSetResult(currentBuild, 'UNSTABLE')}
        //     cleanup {
        //         piperPipelineStagePost script: parameters.script
        //     }
        // }
    }
}
