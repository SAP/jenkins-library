import com.sap.piper.ConfigurationHelper
import com.sap.piper.DebugReport
import com.sap.piper.GenerateDocumentation
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /** Directory where the extensions are cloned to*/
    'globalExtensionsDirectory',
    /** Git url of the repository containing the extensions*/
    'globalExtensionsRepository',
    /** Credentials required to clone the repository*/
    'globalExtensionsRepositoryCredentialsId',
    /** Version which should be each, e.g. the tag name*/
    'globalExtensionsVersion'
]

@Field Set STEP_CONFIG_KEYS = []

@Field Set PARAMETER_KEYS = [
    /** This step will reinitialize the defaults. Make sure to pass the same customDefaults as to the step setupCommonPipelineEnvironment*/
    'customDefaults',
    /** This step will reinitialize the defaults. Make sure to pass the same customDefaultsFromFiles as to the step setupCommonPipelineEnvironment*/
    'customDefaultsFromFiles'
]

/**
 * This step is part of the step setupCommonPipelineEnvironment and should not be used outside independently in a custom pipeline.
 * This step allows users to define extensions (https://sap.github.io/jenkins-library/extensibility/#1-extend-individual-stages) globally instead of in each repository.
 * Instead of defining the extensions in the .pipeline folder the extensions are defined in another repository.
 * You can also place a file called extension_configuration.yml in this repository.
 * Configuration defined in this file will be treated as default values with a lower precedence then custom defaults defined in the project configuration.
 * You can also define additional Jenkins libraries these extensions depend on using a yaml file called sharedLibraries.yml:
 * Example:
 * - name: my-extension-dependency
 *   version: git-tag
 */
@GenerateDocumentation
void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters)
        // load default & individual configuration
        Map configuration = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        echo configuration.toString()
        if(!configuration.globalExtensionsRepository){
            return
        }

        dir(configuration.globalExtensionsDirectory){
            Map gitParameters = [
                $class: 'GitSCM',
                userRemoteConfigs: [[url: configuration.globalExtensionsRepository]]
            ]

            if(configuration.globalExtensionsRepositoryCredentialsId){
                gitParameters.userRemoteConfigs[0].credentialsId = configuration.globalExtensionsRepositoryCredentialsId
            }

            if(configuration.globalExtensionsVersion){
                gitParameters.branches = [[name: configuration.globalExtensionsVersion]]
            }

            checkout(gitParameters)
        }

        String extensionConfigurationFilePath = "${configuration.globalExtensionsDirectory}/extension_configuration.yml"
        if (fileExists(extensionConfigurationFilePath)) {
            writeFile file: ".pipeline/extension_configuration.yml", text: readFile(file: extensionConfigurationFilePath)
            DebugReport.instance.globalExtensionConfigurationFilePath = extensionConfigurationFilePath

            prepareDefaultValues([
                script: script,
                customDefaults: parameters.customDefaults,
                customDefaultsFromFiles: ['extension_configuration.yml'] + parameters.customDefaultsFromFiles
            ])
        }

        def globalExtensionsLibraryConfig = "${configuration.globalExtensionsDirectory}/sharedLibraries.yml"

        if(fileExists(globalExtensionsLibraryConfig)){
            loadLibrariesFromFile(globalExtensionsLibraryConfig)
        }
    }
}

private loadLibrariesFromFile(String filename) {
    List libs = readYaml file: filename
    Set additionalLibraries = []
    for (i = 0; i < libs.size(); i++) {
        Map lib = libs[i]
        String libName = libs[i].name
        String branch = libs[i].version ?: 'master'
        additionalLibraries.add("${libName} | ${branch}")
        library "${libName}@${branch}"
    }
    DebugReport.instance.additionalSharedLibraries.addAll(additionalLibraries)
}
