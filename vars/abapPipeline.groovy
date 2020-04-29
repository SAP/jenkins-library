void call(parameters) {
    node() {
        // agent none
        // //triggers {
        // //    issueCommentTrigger('.*/piper ([a-z]*).*')
        // //}
        // options {
        //     checkoutSCM()
        //     timestamps()
        // }
        // stages {

            stage('Init') {
                steps {
                    checkout scm
            //     library 'piper-lib-os-dev'
            //     //setupCommonPipelineEnvironment script: parameters.script
            //     piperPipelineStageInit script: parameters.script, customDefaults: ['com.sap.piper/pipeline/stageOrdinals.yml'].plus(parameters.customDefaults ?: [])

                    setupCommonPipelineEnvironment script: parameters.script
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
    // }
}
