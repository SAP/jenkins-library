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
                        dockerEnvVars: [GOPATH: '/jenkinsdata/abapPipeline Test/workspace']
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

                    abapEnvironmentPipelineInit script: parameters.script
                }
            }

            stage('Prepare System') {
                steps {
                    cloudFoundryCreateService script: parameters.script
                    input message: "Steampunk system ready?"
                }
            }


            stage('Prepare Communication') {
                steps {
                    cloudFoundryCreateServiceKey script: parameters.script
                }
            }

            stage('Clone Repositories') {
                steps {
                    abapEnvironmentPullGitRepo script: parameters.script
                }
            }


            // stage('Test') {
            //     steps {
            //         abapPipelinePrepare script: parameters.script
            //     }
            // }

            // stage('Delete System') {
            //     steps {
            //         cloudFoundryDeleteService script: parameters.script
            //     }
            // }
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
