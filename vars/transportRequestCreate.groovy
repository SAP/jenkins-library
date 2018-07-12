import com.sap.piper.GitUtils
import groovy.transform.Field

import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException


@Field def STEP_NAME = 'transportRequestCreate'

@Field Set parameterKeys = [
    'changeDocumentId',
    'clientOpts',
    'developmentSystemId',
    'credentialsId',
    'endpoint'
  ]

@Field Set stepConfigurationKeys = [
    'credentialsId',
    'clientOpts',
    'endpoint'
  ]

@Field generalConfigurationKeys = stepConfigurationKeys

def call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        ChangeManagement cm = parameters.cmUtils ?: new ChangeManagement(script)

        Map configuration = ConfigurationHelper
                            .loadStepDefaults(this)
                            .mixinGeneralConfig(script.commonPipelineEnvironment, generalConfigurationKeys)
                            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, stepConfigurationKeys)
                            .mixinStepConfig(script.commonPipelineEnvironment, stepConfigurationKeys)
                            .mixin(parameters, parameterKeys)
                            .withMandatoryProperty('endpoint')
                            .withMandatoryProperty('changeDocumentId')
                            .withMandatoryProperty('developmentSystemId')
                            .use()

        def transportRequestId

        echo "[INFO] Creating transport request for change document '${configuration.changeDocumentId}' and development system '${configuration.developmentSystemId}'."

        withCredentials([usernamePassword(
            credentialsId: configuration.credentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {

            try {
                transportRequestId = cm.createTransportRequest(configuration.changeDocumentId,
                                                               configuration.developmentSystemId,
                                                               configuration.endpoint,
                                                               username,
                                                               password,
                                                               configuration.clientOpts)
            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }
        }

        echo "[INFO] Transport Request '$transportRequestId' has been successfully created."
        return transportRequestId
    }
}
