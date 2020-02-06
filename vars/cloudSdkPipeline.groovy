void call(parameters) {
    final def pipelineSdkVersion = 'master'

    pipeline {
        agent any
        options {
            timeout(time: 120, unit: 'MINUTES')
            timestamps()
            buildDiscarder(logRotator(numToKeepStr: '10', artifactNumToKeepStr: '10'))
            skipDefaultCheckout()
        }
        stages {
            stage('Init') {
                steps {
                    milestone 10
                    library "s4sdk-pipeline-library@${pipelineSdkVersion}"
                    stageInitS4sdkPipeline script: parameters.script
                    abortOldBuilds script: parameters.script
                }
            }

            stage('Build') {
                steps {
                    milestone 20
                    stageBuild script: parameters.script
                }
            }

            stage('Local Tests') {
                parallel {
                    stage("Static Code Checks") {
                        when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.STATIC_CODE_CHECKS } }
                        steps { stageStaticCodeChecks script: parameters.script }
                    }
                    stage("Lint") {
                        when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.LINT } }
                        steps { stageLint script: parameters.script }
                    }
                    stage("Backend Unit Tests") {
                        when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.BACKEND_UNIT_TESTS } }
                        steps { stageUnitTests script: parameters.script }
                    }
                    stage("Backend Integration Tests") {
                        when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.BACKEND_INTEGRATION_TESTS } }
                        steps { stageBackendIntegrationTests script: parameters.script }
                    }
                    stage("Frontend Integration Tests") {
                        when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.FRONTEND_INTEGRATION_TESTS } }
                        steps { stageFrontendIntegrationTests script: parameters.script }
                    }
                    stage("Frontend Unit Tests") {
                        when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.FRONTEND_UNIT_TESTS } }
                        steps { stageFrontendUnitTests script: parameters.script }
                    }
                    stage("NPM Dependency Audit") {
                        when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.NPM_AUDIT } }
                        steps { stageNpmAudit script: parameters.script }
                    }
                }
            }

            stage('Remote Tests') {
                when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.REMOTE_TESTS } }
                parallel {
                    stage("End to End Tests") {
                        when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.E2E_TESTS } }
                        steps { stageEndToEndTests script: parameters.script }
                    }
                    stage("Performance Tests") {
                        when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.PERFORMANCE_TESTS } }
                        steps { stagePerformanceTests script: parameters.script }
                    }
                }
            }

            stage('Quality Checks') {
                when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.QUALITY_CHECKS } }
                steps {
                    milestone 50
                    stageS4SdkQualityChecks script: parameters.script
                }
            }

            stage('Third-party Checks') {
                when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.THIRD_PARTY_CHECKS } }
                parallel {
                    stage("Checkmarx Scan") {
                        when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.CHECKMARX_SCAN } }
                        steps { stageCheckmarxScan script: parameters.script }
                    }
                    stage("WhiteSource Scan") {
                        when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.WHITESOURCE_SCAN } }
                        steps { stageWhitesourceScan script: parameters.script }
                    }
                    stage("SourceClear Scan") {
                        when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.SOURCE_CLEAR_SCAN } }
                        steps { stageSourceClearScan script: parameters.script }
                    }
                    stage("Fortify Scan") {
                        when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.FORTIFY_SCAN } }
                        steps { stageFortifyScan script: parameters.script }
                    }
                    stage("Additional Tools") {
                        when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.ADDITIONAL_TOOLS } }
                        steps { stageAdditionalTools script: parameters.script }
                    }
                    stage('SonarQube Scan') {
                        when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.SONARQUBE_SCAN } }
                        steps { stageSonarQubeScan script: parameters.script }
                    }
                }
            }

            stage('Artifact Deployment') {
                when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.ARTIFACT_DEPLOYMENT } }
                steps {
                    milestone 70
                    stageArtifactDeployment script: parameters.script
                }
            }

            stage('Production Deployment') {
                when { expression { parameters.script.commonPipelineEnvironment.configuration.runStage.PRODUCTION_DEPLOYMENT } }
                //milestone 80 is set in stageProductionDeployment
                steps { stageProductionDeployment script: parameters.script }
            }

        }
        post {
            always {
                script {
                    postActionArchiveDebugLog script: parameters.script
                    if (parameters.script.commonPipelineEnvironment?.configuration?.runStage?.SEND_NOTIFICATION) {
                        postActionSendNotification script: parameters.script
                    }
                    postActionCleanupStashesLocks script: parameters.script
                    sendAnalytics script: parameters.script

                    if (parameters.script.commonPipelineEnvironment?.configuration?.runStage?.POST_PIPELINE_HOOK) {
                        stage('Post Pipeline Hook') {
                            stagePostPipelineHook script: parameters.script
                        }
                    }
                }
            }
            success {
                script {
                    if (parameters.script.commonPipelineEnvironment?.configuration?.runStage?.ARCHIVE_REPORT) {
                        postActionArchiveReport script: parameters.script
                    }
                }
            }
            failure { deleteDir() }
        }
    }
}
