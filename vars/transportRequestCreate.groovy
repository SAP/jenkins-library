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

@Field def STEP_NAME = 'transportRequestCreate'

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
            /**
              * Where/how the transport request is created (via SAP Solution Manager, ABAP).
              *  @possibleValues `SOLMAN`, `CTS`, `NONE`
              */
            .withMandatoryProperty('changeManagement/type')


        Map configuration =  configHelper.use()

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
              * The starting point for retrieving the change document id.
              */
            .withMandatoryProperty('changeManagement/git/from')
            /**
              * The end point for retrieving the change document id.
              */
            .withMandatoryProperty('changeManagement/git/to')
            /**
              * Specifies what part of the commit is scanned. By default the body of the commit message is scanned.
              * @possibleValues see `git log --help`
              */
            .withMandatoryProperty('changeManagement/git/format')
            /**
              * For type `SOLMAN` only. A pattern used for identifying lines holding the change document id.
              * @possibleValues regex pattern
              * @mandatory `SOLMAN` only
              */
            .withMandatoryProperty('changeManagement/changeDocumentLabel', null, { backendType == BackendType.SOLMAN})
            /**
              * for type `CTS` only. Typically `W` (workbench) or `C` customizing.
              * @mandatory `CTS` only
              */
            .withMandatoryProperty('transportType', null, { backendType == BackendType.CTS})
            /**
              * For type `CTS` only. The system receiving the transport request.
              * @mandatory `CTS` only
              */
            .withMandatoryProperty('targetSystem', null, { backendType == BackendType.CTS})
            /**
              * For type `CTS` only. The description of the transport request.
              * @mandatory `CTS` only
              */
            .withMandatoryProperty('description', null, { backendType == BackendType.CTS})

        def changeDocumentId = null

        new Utils().pushToSWA([step: STEP_NAME,
                                stepParam1: parameters?.script == null], configuration)

        if(backendType == BackendType.SOLMAN) {

            changeDocumentId = getChangeDocumentId(cm, script, configuration)

            configHelper.mixin([changeDocumentId: changeDocumentId?.trim() ?: null], ['changeDocumentId'] as Set)
                        /**
                          * for `SOLMAN` only. Outlines how the artifact is handled.
                          * For CTS use case: `SID~Type/Client`, e.g. `XX1~ABAP/100`, for SOLMAN use case: `SID~Typ`,
                          * e.g. `J01~JAVA`.
                          * @mandatory `SOLMAN` only.
                          */
                        .withMandatoryProperty('developmentSystemId')
                        /**
                          * for `SOLMAN` only. The id of the change document to that the transport request is bound to.
                          * Typically this value is provided via commit message in the commit history.
                          * @mandatory `SOLMAN` only, can be provided via git commit history.
                          */
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
