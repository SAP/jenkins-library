import com.sap.piper.GitUtils
import groovy.transform.Field

import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException


@Field def STEP_NAME = 'transportRequestCreate'

@Field Set stepConfigurationKeys = [
    'credentialsId',
    'clientOpts',
    'endpoint',
    'gitFrom',
    'gitTo',
    'gitChangeDocumentLabel',
    'gitFormat'
  ]

@Field Set parameterKeys = stepConfigurationKeys.plus(['changeDocumentId', 'developmentSystemId'])

@Field generalConfigurationKeys = stepConfigurationKeys

def call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        ChangeManagement cm = parameters.cmUtils ?: new ChangeManagement(script)

        ConfigurationHelper configHelper = ConfigurationHelper
                                            .loadStepDefaults(this)
                                            .mixinGeneralConfig(script.commonPipelineEnvironment, generalConfigurationKeys)
                                            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, stepConfigurationKeys)
                                            .mixinStepConfig(script.commonPipelineEnvironment, stepConfigurationKeys)
                                            .mixin(parameters, parameterKeys)
                                            .withMandatoryProperty('endpoint')
                                            .withMandatoryProperty('developmentSystemId')

        Map configuration =  configHelper.use()

        def changeDocumentId = configuration.changeDocumentId

        if(changeDocumentId?.trim()) {

            echo "[INFO] ChangeDocumentId '${changeDocumentId}' retrieved from parameters."

        } else {

            echo "[INFO] Retrieving ChangeDocumentId from commit history [from: ${configuration.gitFrom}, to: ${configuration.gitTo}]." +
                 "Searching for pattern '${configuration.gitChangeDocumentLabel}'. Searching with format '${configuration.gitFormat}'."

            try {
                changeDocumentId = cm.getChangeDocumentId(
                                                          configuration.gitFrom,
                                                          configuration.gitTo,
                                                          configuration.gitChangeDocumentLabel,
                                                          configuration.gitFormat
                                                         )

                echo "[INFO] ChangeDocumentId '${changeDocumentId}' retrieved from commit history"
            } catch(ChangeManagementException ex) {
                echo "[WARN] Cannot retrieve changeDocumentId from commit history: ${ex.getMessage()}."
            }
        }

        configuration = configHelper.mixin([changeDocumentId: changeDocumentId?.trim() ?: null], ['changeDocumentId'] as Set)
                                    .withMandatoryProperty('changeDocumentId',
                                        "Change document id not provided (parameter: \'changeDocumentId\' or via commit history).")
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
