import com.sap.piper.ConfigurationHelper
import com.sap.piper.DebugReport
import com.sap.piper.GenerateDocumentation
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /** Credentials (username and password) used to download custom defaults if access is secured.*/
    'globalExtensionsDirectory',
    /** Credentials (username and password) used to download custom defaults if access is secured.*/
    'globalExtensionsRepository',
    /** Credentials (username and password) used to download custom defaults if access is secured.*/
    'globalExtensionsRepositoryCredentialsId',
    /** Credentials (username and password) used to download custom defaults if access is secured.*/
    'globalExtensionsVersion'
]

@Field Set STEP_CONFIG_KEYS = []

@Field Set PARAMETER_KEYS = [
    /** Credentials (username and password) used to download custom defaults if access is secured.*/
    'customDefaults',
    /** Credentials (username and password) used to download custom defaults if access is secured.*/
    'customDefaultsFromFiles'
]

/**
 *
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
            parameters.customDefaultsFromFiles = [ extensionConfigurationFilePath ]

            prepareDefaultValues([
                script: script,
                customDefaults: parameters.customDefaults,
                customDefaultsFromFiles: ['extension_configuration.yml'] + parameters.customDefaultsFromFiles
            ])
        }
    }
}
