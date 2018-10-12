import com.sap.piper.Utils

import groovy.transform.Field


@Field STEP_NAME = 'pipelineExecute'


/**
 * pipelineExecute
 * Load and executes a pipeline from another git repository.
 *
 */
def call(Map parameters = [:]) {

    node() {

        def path

        handlePipelineStepErrors (stepName: 'pipelineExecute', stepParameters: parameters) {

            def utils = new Utils()

            // The coordinates of the pipeline script
            def repo = utils.getMandatoryParameter(parameters, 'repoUrl', null)
            def branch = utils.getMandatoryParameter(parameters, 'branch', 'master')

            path = utils.getMandatoryParameter(parameters, 'path', 'Jenkinsfile')

            // In case access to the repository containing the pipeline
            // script is restricted the credentialsId of the credentials used for
            // accessing the repository needs to be provided below. The corresponding
            // credentials needs to be configured in Jenkins accordingly.
            def credentialsId = utils.getMandatoryParameter(parameters, 'credentialsId', '')

            deleteDir()

            checkout([$class: 'GitSCM', branches: [[name: branch]],
                      doGenerateSubmoduleConfigurations: false,
                      extensions: [[$class: 'SparseCheckoutPaths',
                                    sparseCheckoutPaths: [[path: path]]
                                   ]],
                      submoduleCfg: [],
                      userRemoteConfigs: [[credentialsId: credentialsId,
                                           url: repo
                                          ]]
            ])

        }
        load path
    }
}
