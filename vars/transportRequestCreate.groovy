import com.sap.piper.Utils
import groovy.transform.Field

import com.sap.piper.ConfigurationHelper
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException


@Field def STEP_NAME = 'transportRequestCreate'

@Field Set stepConfigurationKeys = [
    'changeManagement',
     'developmentSystemId'
  ]

@Field Set parameterKeys = stepConfigurationKeys.plus(['changeDocumentId'])

@Field generalConfigurationKeys = stepConfigurationKeys

def call(parameters = [:]) {

    def transportRequestId

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        ChangeManagement cm = parameters.cmUtils ?: new ChangeManagement(script)

        ConfigurationHelper configHelper = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, generalConfigurationKeys)
            .mixinStepConfig(script.commonPipelineEnvironment, stepConfigurationKeys)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, stepConfigurationKeys)
            .mixin(parameters, parameterKeys)
            .withMandatoryProperty('changeManagement/clientOpts')
            .withMandatoryProperty('changeManagement/credentialsId')
            .withMandatoryProperty('changeManagement/endpoint')
            .withMandatoryProperty('changeManagement/git/from')
            .withMandatoryProperty('changeManagement/git/to')
            .withMandatoryProperty('changeManagement/git/format')
            .withMandatoryProperty('developmentSystemId')

        Map configuration =  configHelper.use()

        new Utils().pushToSWA([step: STEP_NAME], configuration)

        def changeDocumentId = configuration.changeDocumentId

        if(changeDocumentId?.trim()) {

            echo "[INFO] ChangeDocumentId '${changeDocumentId}' retrieved from parameters."

        } else {

            echo "[INFO] Retrieving ChangeDocumentId from commit history [from: ${configuration.changeManagement.git.from}, to: ${configuration.changeManagement.git.to}]." +
                 "Searching for pattern '${configuration.changeDocumentLabel}'. Searching with format '${configuration.changeManagement.git.format}'."

            try {

                changeDocumentId = cm.getChangeDocumentId(
                                                          configuration.changeManagement.git.from,
                                                          configuration.changeManagement.git.to,
                                                          configuration.changeManagement.changeDocumentLabel,
                                                          configuration.changeManagement.git.format
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

        echo "[INFO] Creating transport request for change document '${configuration.changeDocumentId}' and development system '${configuration.developmentSystemId}'."

            try {
                transportRequestId = cm.createTransportRequest(configuration.changeDocumentId,
                                                               configuration.developmentSystemId,
                                                               configuration.changeManagement.endpoint,
                                                               configuration.changeManagement.credentialsId,
                                                               configuration.changeManagement.clientOpts)
            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }


        echo "[INFO] Transport Request '$transportRequestId' has been successfully created."
    }

    return transportRequestId
}
