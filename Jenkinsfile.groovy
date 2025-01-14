pipeline {
    agent any
    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        stage('Execute Custom Step') {
            steps {
                script {
                    piperTestStep(message: 'Testing Onapsis integration')
                }
            }
        }
    }
}