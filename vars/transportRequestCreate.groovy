import com.sap.piper.GitUtils
import com.sap.piper.Utils
import groovy.transform.Field

import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.BackendType
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import static com.sap.piper.cm.StepHelpers.getBackendTypeAndLogInfoIfCMIntegrationDisabled

import hudson.AbortException

@Field def STEP_NAME = 'transportRequestCreate'

@Field Set stepConfigurationKeys = [
    'changeManagement',
    'description',          // CTS
    'developmentSystemId',  // SOLMAN
    'targetSystem',         // CTS
    'transportType',        // CTS
  ]

@Field Set parameterKeys = stepConfigurationKeys.plus(['changeDocumentId'])

@Field generalConfigurationKeys = stepConfigurationKeys

def call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        ChangeManagement cm = parameters.cmUtils ?: new ChangeManagement(script)

        ConfigurationHelper configHelper = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, generalConfigurationKeys)
            .mixinStepConfig(script.commonPipelineEnvironment, stepConfigurationKeys)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, stepConfigurationKeys)
            .mixin(parameters, parameterKeys)


        Map configuration =  configHelper.use()

        BackendType backendType = getBackendTypeAndLogInfoIfCMIntegrationDisabled(this, configuration)
        if(backendType == BackendType.NONE) return

        new Utils().pushToSWA([step: STEP_NAME], configuration)

        configHelper
            .withMandatoryProperty('changeManagement/clientOpts')
            .withMandatoryProperty('changeManagement/credentialsId')
            .withMandatoryProperty('changeManagement/endpoint')
            .withMandatoryProperty('changeManagement/git/from')
            .withMandatoryProperty('changeManagement/git/to')
            .withMandatoryProperty('changeManagement/git/format')
            .withMandatoryProperty('transportType', null, { backendType == BackendType.CTS})
            .withMandatoryProperty('targetSystem', null, { backendType == BackendType.CTS})
            .withMandatoryProperty('description', null, { backendType == BackendType.CTS})

        def changeDocumentId = null

        if(backendType == BackendType.SOLMAN) {

            changeDocumentId = configuration.changeDocumentId

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
            configHelper.mixin([changeDocumentId: changeDocumentId?.trim() ?: null], ['changeDocumentId'] as Set)
                        .withMandatoryProperty('developmentSystemId')
                        .withMandatoryProperty('changeDocumentId',
                            "Change document id not provided (parameter: \'changeDocumentId\' or via commit history).")
        }

        configuration = configHelper.use()

        def transportRequestId

        def creatingMessage = ["[INFO] Creating transport request"]
        if(backendType == BackendType.SOLMAN) {
            creatingMessage << " for change document '${configuration.changeDocumentId}' and development system '${configuration.developmentSystemId}'"
        }
        creatingMessage << '.'
        echo creatingMessage.join()

            try {
                if(backendType == BackendType.SOLMAN) {
                    transportRequestId = cm.createTransportRequestSOLMAN(
                                                               configuration.changeDocumentId,
                                                               configuration.developmentSystemId,
                                                               configuration.changeManagement.endpoint,
                                                               configuration.changeManagement.credentialsId,
                                                               configuration.changeManagement.clientOpts)
                } else if(backendType == BackendType.CTS) {
                    transportRequestId = cm.createTransportRequestCTS(
                                                               configuration.transportType,
                                                               configuration.targetSystem,
                                                               configuration.description,
                                                               configuration.changeManagement.endpoint,
                                                               configuration.changeManagement.credentialsId,
                                                               configuration.changeManagement.clientOpts)
                } else {
                  throw new IllegalArgumentException("Invalid backend type: '${backendType}'.")
                }
            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }


        echo "[INFO] Transport Request '$transportRequestId' has been successfully created."
        return transportRequestId
    }
}
