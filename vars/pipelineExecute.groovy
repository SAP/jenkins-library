import com.sap.piper.Utils

import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger


/**
 * pipelineExecute
 * Load and executes a pipeline from another git repository.
 *
 */
def call(Map parameters = [:]) {

    node() {

        def path

        def stepName = 'pipelineExecute'

        List parameterKeys = [
            'repoUrl',
            'branch',
            'path',
            'credentialsId'
        ]

        List stepConfigurationKeys = [
            'repoUrl',
            'branch',
            'path',
            'credentialsId'
        ]

        handlePipelineStepErrors (stepName: stepName, stepParameters: parameters) {

            final script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

            prepareDefaultValues script: script

            final Map stepConfiguration = ConfigurationLoader.stepConfiguration(script, stepName)
            final Map stepDefaults = ConfigurationLoader.defaultStepConfiguration(script, stepName)
            final Map configuration = ConfigurationMerger.merge(parameters, parameterKeys,
                                      stepConfiguration, stepConfigurationKeys, stepDefaults)

            def repo = configuration.repoUrl
            def branch = configuration.branch
            path = configuration.path
            def credentialsId = configuration.credentialsId

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
