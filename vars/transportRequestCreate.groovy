import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.Utils
import groovy.transform.Field

import com.sap.piper.ConfigurationHelper
import com.sap.piper.cm.BackendType
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import static com.sap.piper.cm.StepHelpers.getBackendTypeAndLogInfoIfCMIntegrationDisabled

import static com.sap.piper.cm.StepHelpers.getChangeDocumentId
import hudson.AbortException

@Field def STEP_NAME = getClass().getName()

@Field GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

@Field Set STEP_CONFIG_KEYS = [
    'changeManagement',
    'description',          // CTS
    'developmentSystemId',  // SOLMAN
    'targetSystem',         // CTS
    'transportType',        // CTS
  ]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus(['changeDocumentId'])

void call(parameters = [:]) {

    def transportRequestId

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this

        ChangeManagement cm = parameters.cmUtils ?: new ChangeManagement(script)

        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)


        Map configuration =  configHelper.use()

        BackendType backendType = getBackendTypeAndLogInfoIfCMIntegrationDisabled(this, configuration)
        if(backendType == BackendType.NONE) return

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
            .withMandatoryProperty('changeManagement/rfc/client', null, {backendType == BackendType.RFC})

        def changeDocumentId = null

        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], configuration)

        if(backendType == BackendType.SOLMAN) {

            changeDocumentId = getChangeDocumentId(cm, script, configuration)

            configHelper.mixin([changeDocumentId: changeDocumentId?.trim() ?: null], ['changeDocumentId'] as Set)
                        .withMandatoryProperty('developmentSystemId')
                        .withMandatoryProperty('changeDocumentId',
                            "Change document id not provided (parameter: \'changeDocumentId\' or via commit history).")
        }

        configuration = configHelper.use()

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
                } else if (backendType == BackendType.RFC) {
                  transportRequestId = cm.createTransportRequestRFC(
                                                               configuration.changeManagement.rfc.dockerImage,
                                                               configuration.changeManagement.rfc.dockerOptions ?: [],
                                                               configuration.changeManagement.endpoint,
                                                               configuration.changeManagement.rfc.client,
                                                               configuration.changeManagement.credentialsId,
                                                               configuration.description)
                } else {
                  throw new IllegalArgumentException("Invalid backend type: '${backendType}'.")
                }
            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }


        echo "[INFO] Transport Request '$transportRequestId' has been successfully created."
        script.commonPipelineEnvironment.setTransportRequestId(transportRequestId)
    }
}
