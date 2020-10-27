import static com.sap.piper.Prerequisites.checkScript

import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()

@Field def GENERAL_CONFIG_KEYS = []
@Field def PARAMETER_KEYS = []
@Field def STEP_CONFIG_KEYS = []

/** The Scenario is intended for building and uploading a fiori application.
  *
  * It needs to be called from a pipeline script (Jenkinsfile) like:
  * ```
  *   @Library('piper-lib-os') _
  *   @Library('your-additional-lib') __ // optional
  *
  *   // parameter 'customDefaults' below is optional
  *   fioriOnCloudPlatformPipeline(script: this, customDefaults: '<configFile>')
  * ```
  */
void call(parameters = [:]) {

    checkScript(this, parameters)

    node(parameters.label) {

        //
        // Cut and paste lines below in order to create a pipeline from this scenario
        // In this case `parameters` needs to be replaced by `script: this`.

        stage('prepare') {

            deleteDir()
            checkout scm
            setupCommonPipelineEnvironment(parameters)
        }

        stage('build') {

            mtaBuild(parameters)
        }

        stage('deploy') {

            def mtaBuildCfg = parameters.script.commonPipelineEnvironment.getStepConfiguration('mtaBuild', '')

            if((mtaBuildCfg.platform == 'NEO') || (mtaBuildCfg.buildTarget == 'NEO')) {
                neoDeploy(parameters)
            }
            else if((mtaBuildCfg.platform == 'CF') || (mtaBuildCfg.buildTarget == 'CF')) {
                cloudFoundryDeploy(parameters)
            }
            else {
                error "Deployment failed: no valid deployment target defined! Find details in https://sap.github.io/jenkins-library/steps/mtaBuild/#platform"
            }
        }

        // Cut and paste lines above in order to create a pipeline from this scenario
        //
    }
}
