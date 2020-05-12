void call(parameters) {
    pipeline {
        agent any
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
                    abapEnvironmentPipelineInit script: parameters.script
                }
            }

            stage('Prepare') {
                steps {
                    cloudFoundryCreateService script: parameters.script
                    input message: "Steampunk system ready?"
                    cloudFoundryCreateServiceKey script: parameters.script
                }
            }

            stage('Clone') {
                steps {
                    abapEnvironmentPullGitRepo script: parameters.script
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
                cloudFoundryDeleteService script: parameters.script
            }
        }
    }
}
