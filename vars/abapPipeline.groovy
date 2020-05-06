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
            stage("go binary"){
                steps {
                    sh '''
                        rm -rf jenkins-library
                        git clone https://github.com/DanielMieg/jenkins-library.git
                    '''

                    dockerExecute(
                        script: this,
                        dockerImage: 'golang',
                        dockerEnvVars: [GOPATH: '/jenkinsdata/cloudFoundryDeleteService Test Pipeline/workspace']
                    ) {
                        sh '''
                            cd jenkins-library
                            go build -o piper .
                            chmod +x piper
                            cp piper ..
                            cd ..
                        '''
                        stash name: 'piper-bin', includes: 'piper'
                    }
                }
            }
            stage('Init') {
                steps {
            //     library 'piper-lib-os-dev'
            //     //setupCommonPipelineEnvironment script: parameters.script
            //     piperPipelineStageInit script: parameters.script, customDefaults: ['com.sap.piper/pipeline/stageOrdinals.yml'].plus(parameters.customDefaults ?: [])

                    abapPipelineInit script: parameters.script
                }
            }

            stage('Prepare') {
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
