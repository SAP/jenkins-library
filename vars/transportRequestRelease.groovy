import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.Utils
import groovy.transform.Field

import com.sap.piper.ConfigurationHelper
import com.sap.piper.cm.BackendType
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException

import static com.sap.piper.cm.StepHelpers.getTransportRequestId
import static com.sap.piper.cm.StepHelpers.getChangeDocumentId
import static com.sap.piper.cm.StepHelpers.getBackendTypeAndLogInfoIfCMIntegrationDisabled

@Field def STEP_NAME = 'transportRequestRelease'

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

@Field Set STEP_CONFIG_KEYS = [
    'changeManagement'
  ]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /**
      * for `SOLMAN` only. The id of the change document related to the transport request to release.
      * @mandatory `SOLMAN` only
      */
    'changeDocumentId',
    /**
      * The id of the transport request to release.
      */
    'transportRequestId',
  ])

void call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this

        ChangeManagement cm = parameters.cmUtils ?: new ChangeManagement(script)

        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            /**
              * Where/how the transport request is created (via SAP Solution Manager, ABAP).
              * @possibleValues `SOLMAN`, `CTS`, `NONE`
              */
            .withMandatoryProperty('changeManagement/type')


        Map configuration = configHelper.use()

        BackendType backendType = getBackendTypeAndLogInfoIfCMIntegrationDisabled(this, configuration)
        if(backendType == BackendType.NONE) return

        configHelper
            /**
              * Options forwarded to JVM used by the CM client, like `JAVA_OPTS`.
              */
            .withMandatoryProperty('changeManagement/clientOpts')
            /**
              * The credentials to connect to the service endpoint (Solution Manager, ABAP System).
              */
            .withMandatoryProperty('changeManagement/credentialsId')
            /**
              * The service endpoint (Solution Manager, ABAP System).
              */
            .withMandatoryProperty('changeManagement/endpoint')
            /**
              * The end point for retrieving the change document id and/or transport request id.
              */
            .withMandatoryProperty('changeManagement/git/to')
            /**
              * The starting point for retrieving the change document id and/or transport request id
              */
            .withMandatoryProperty('changeManagement/git/from')
            /**
              * Specifies what part of the commit is scanned. By default the body of the commit message is scanned.
              */
            .withMandatoryProperty('changeManagement/git/format')
            /**
              * For type `SOLMAN` only. A pattern used for identifying lines holding the change document id.
              * @possibleValues regex pattern
              */
            .withMandatoryProperty('changeManagement/changeDocumentLabel')
            /**
              * A pattern used for identifying lines holding the transport request id.
              * @possibleValues regex pattern
              */
            .withMandatoryProperty('changeManagement/transportRequestLabel')

        configuration = configHelper.use()

        new Utils().pushToSWA([step: STEP_NAME,
                                stepParam1: parameters?.script == null], configuration)

        def changeDocumentId = null
        def transportRequestId = getTransportRequestId(cm, script, configuration)

        if(backendType == BackendType.SOLMAN) {

            changeDocumentId = getChangeDocumentId(cm, script, configuration)

            configHelper.mixin([changeDocumentId: changeDocumentId?.trim() ?: null], ['changeDocumentId'] as Set)
                        .withMandatoryProperty('changeDocumentId',
                            "Change document id not provided (parameter: \'changeDocumentId\' or via commit history).")

        }

        configuration = configHelper
                            .mixin([transportRequestId: transportRequestId?.trim() ?: null], ['transportRequestId'] as Set)
                            .withMandatoryProperty('transportRequestId',
                                "Transport request id not provided (parameter: \'transportRequestId\' or via commit history).")
                            .use()

        def closingMessage = ["[INFO] Closing transport request '${configuration.transportRequestId}'"]
        if(backendType == BackendType.SOLMAN) closingMessage << " for change document '${configuration.changeDocumentId}'"
        closingMessage << '.'
        echo closingMessage.join()

            try {
                cm.releaseTransportRequest(backendType,
                                           configuration.changeDocumentId,
                                           configuration.transportRequestId,
                                           configuration.changeManagement.endpoint,
                                           configuration.changeManagement.credentialsId,
                                           configuration.changeManagement.clientOpts)

            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }


        echo "[INFO] Transport Request '${configuration.transportRequestId}' has been successfully closed."
    }
}
