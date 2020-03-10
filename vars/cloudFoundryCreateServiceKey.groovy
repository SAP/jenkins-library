import com.sap.piper.GenerateDocumentation
import com.sap.piper.BashUtils
import com.sap.piper.PiperGoUtils
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field def STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/cloudFoundryCreateServiceKey.yaml'

@Field Set STEP_CONFIG_KEYS = [
    'cloudFoundry',
        /**
         * Cloud Foundry API endpoint.
         * @parentConfigKey cloudFoundry
         */
        'apiEndpoint',
        /**
         * Cloud Foundry credentials.
         * @parentConfigKey cloudFoundry
         */
        'credentialsId',
        /**
         * Cloud Foundry target organization.
         * @parentConfigKey cloudFoundry
         */
        'org',
        /**
         * Cloud Foundry target space.
         * @parentConfigKey cloudFoundry
         */
        'space',
        /**
         * Cloud Foundry service instance, for which the service key will be created.
         * @parentConfigKey cloudFoundry
         */
        'serviceInstance',
        /**
         * Cloud Foundry service key, which will be created.
         * @parentConfigKey cloudFoundry
         */
        'serviceKeyName',
        /**
         * Cloud Foundry service key configuration.
         * @parentConfigKey cloudFoundry
         */
        'serviceKeyConfig',
    /** @see dockerExecute */
    'dockerImage',
    /** @see dockerExecute */
    'dockerWorkspace'
]

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
@Field Map CONFIG_KEY_COMPATIBILITY = [cloudFoundry: [apiEndpoint: 'cfApiEndpoint', credentialsId: 'cfCredentialsId', org: 'cfOrg', space: 'cfSpace', serviceInstance: 'cfServiceInstance']]


void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters, failOnError: true) {

        def script = checkScript(this, parameters) ?: this

        //Map config
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixin(parameters, PARAMETER_KEYS, CONFIG_KEY_COMPATIBILITY)
            .withMandatoryProperty('cloudFoundry/org')
            .withMandatoryProperty('cloudFoundry/space')
            .withMandatoryProperty('cloudFoundry/credentialsId')
            .withMandatoryProperty('cloudFoundry/serviceInstance')
            //.withMandatoryProperty('cloudFoundry/serviceKeyName')
            .withMandatoryProperty('cloudFoundry/apiEndpoint')
            .use()
        def utils = parameters.juStabUtils ?: new Utils()

        script.commonPipelineEnvironment.writeToDisk(script)

        //writeFile(file: METADATA_FILE, text: libraryResource(METADATA_FILE))

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(parameters)}",
        ]) {
            // get context configuration
            config = readJSON (text: sh(returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '${METADATA_FILE}'"))
            // execute step
            dockerExecute(
                script: script,
                dockerImage: config.dockerImage,
                dockerWorkspace: config.dockerWorkspace
            ) {
                withCredentials([usernamePassword(
                    credentialsId: config.cfCredentialsId,
                    passwordVariable: 'PIPER_password',
                    usernameVariable: 'PIPER_username'
                )]) {
                    sh "./piper cloudFoundryCreateServiceKey"
                }
            }
        }
    }
}
