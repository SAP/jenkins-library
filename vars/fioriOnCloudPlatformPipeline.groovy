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
  *   fioriOnCloudPlatformPipeline(customDefaults: '<configFile>')
  * ```
  */
void call(parameters = [:]) {

    node(parameters.label) {

        //
        // Cut and paste lines below in order to create a pipeline from this scenario

        stage('prepare') {

            deleteDir()
            checkout scm
            setupCommonPipelineEnvironment(parameters)
        }

        stage('build') {

            mtaBuild(parameters)
        }

        stage('deploy') {

            neoDeploy(parameters)
        }

        // Cut and paste lines above in order to create a pipeline from this scenario
        //
    }
}
