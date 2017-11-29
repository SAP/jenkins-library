import com.sap.piper.Utils

/**
 * centralPipelineLoad
 * Load a central pipeline.
 *
 */
def call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: 'centralPipelineLoad', stepParameters: parameters) {

        def utils = new Utils()

        // The coordinates of the central pipeline script
        def repo = utils.getMandatoryParameter(parameters, 'repoUrl', null)
        def branch = utils.getMandatoryParameter(parameters, 'branch', 'master')
        def path = utils.getMandatoryParameter(parameters, 'path', 'Jenkinsfile')

        // In case access to the repository containing the central pipeline
        // script is restricted the credentialsId of the credentials used for
        // accessing the repository needs to be provided below. The corresponding
        // credentials needs to be configured in Jenkins accordingly.
        def credentialsId = utils.getMandatoryParameter(parameters, 'credentialsId', '')

        node() {
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

            load path
        }
    }
}
