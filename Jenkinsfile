node {
    try {
        lock(resource: "sap-jenkins-library/10", inversePrecedence: true) {
            milestone 10
            deleteDir()
            stage ('Checkout'){
                checkout scm
            }
            stage ('Test') {
                sh "mvn clean test --batch-mode"
            }
        }
    } catch (Throwable err) {
        echo "Error occured: ${err}"
        currentBuild.result = 'FAILURE'
        mail subject: '[Build failed] SAP/jenkins-library', body: 'Fix the build.', to: 'marcus.holl@sap.com,oliver.nocon@sap.com'
        throw err
    } 
}
