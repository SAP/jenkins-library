import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS.plus([
    /** Credentials (username and password) used to download custom defaults if access is secured.*/
    'globalExtensionsDirectory',
    /** Credentials (username and password) used to download custom defaults if access is secured.*/
    'globalExtensionsRepository',
    /** Credentials (username and password) used to download custom defaults if access is secured.*/
    'globalExtensionsRepositoryCredentialsId',
    /** Credentials (username and password) used to download custom defaults if access is secured.*/
    'globalExtensionsVersion'
])

@Field Set STEP_CONFIG_KEYS = []

@Field Set PARAMETER_KEYS = []

/**
 *
 */
@GenerateDocumentation
void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters)
        // load default & individual configuration
        Map configuration = ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        if(!configuration.globalExtensionsRepository){
            return
        }

        dir(configuration.globalExtensionsDirectory){
            Map gitParameters = [url: configuration.globalExtensionsRepository]
            if(configuration.globalExtensionsRepositoryCredentialsId){
                gitParameters.credentialsId = configuration.globalExtensionsRepositoryCredentialsId
            }
            if(configuration.globalExtensionsVersion){
                gitParameters.branch = configuration.globalExtensionsVersion
            }
        }
    }
}
