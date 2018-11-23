import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.Utils
import groovy.transform.Field

import com.sap.piper.ConfigurationHelper
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.BackendType
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException

import static com.sap.piper.cm.StepHelpers.getTransportRequestId
import static com.sap.piper.cm.StepHelpers.getChangeDocumentId
import static com.sap.piper.cm.StepHelpers.getBackendTypeAndLogInfoIfCMIntegrationDisabled

@Field def STEP_NAME = 'transportRequestUploadFile'

@Field Set GENERAL_CONFIG_KEYS = [
    'changeManagement'
  ]

@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
     /**
       * For type `SOLMAN` only. The id of the application.
       * @mandatory `SOLMAN` only
       */
      'applicationId'
    ])

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /**
      * For type `SOLMAN` only. The id of the change document to that the
      * transport request is bound to. Typically this value is provided via commit message in the commit history.
      * @mandatory `SOLMAN` only
      */
    'changeDocumentId',
    /**
      * The path of the file to upload.
      */
    'filePath',
    /**
      * The id of the transport request to release. Typically provided via commit history.
      */
    'transportRequestId'])

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
            .addIfEmpty('filePath', script.commonPipelineEnvironment.getMtarFilePath())

        Map configuration = configHelper.use()

        BackendType backendType = getBackendTypeAndLogInfoIfCMIntegrationDisabled(this, configuration)
        if(backendType == BackendType.NONE) return

        configHelper
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
              * Where/how the transport request is created (via SAP Solution Manager, ABAP).
              * @possibleValues `SOLMAN`, `CTS`, `NONE`
              */
            .withMandatoryProperty('changeManagement/type')
            /**
              * The starting point for retrieving the change document id and/or transport request id.
              */
            .withMandatoryProperty('changeManagement/git/from')
            /**
              * The end point for retrieving the change document id and/or transport request id.
              */
            .withMandatoryProperty('changeManagement/git/to')
            /**
              * Specifies what part of the commit is scanned. By default the body of the commit message is scanned.
              * @possibleValues see `git log --help`
              */
            .withMandatoryProperty('changeManagement/git/format')
            .withMandatoryProperty('filePath')

        new Utils().pushToSWA([step: STEP_NAME,
                                stepParam1: configuration.changeManagement.type,
                                stepParam2: parameters?.script == null], configuration)

        def changeDocumentId = null

        if(backendType == BackendType.SOLMAN) {
            changeDocumentId = getChangeDocumentId(cm, script, configuration)
        }

        def transportRequestId = getTransportRequestId(cm, script, configuration)

        configHelper
            .mixin([changeDocumentId: changeDocumentId?.trim() ?: null,
                    transportRequestId: transportRequestId?.trim() ?: null], ['changeDocumentId', 'transportRequestId'] as Set)

        if(backendType == BackendType.SOLMAN) {
            configHelper
                .withMandatoryProperty('changeDocumentId',
                    "Change document id not provided (parameter: \'changeDocumentId\' or via commit history).")
                .withMandatoryProperty('applicationId')
        }
        configuration = configHelper
                            .withMandatoryProperty('transportRequestId',
                               "Transport request id not provided (parameter: \'transportRequestId\' or via commit history).")
                           .use()

        def uploadingMessage = ["[INFO] Uploading file '${configuration.filePath}' to transport request '${configuration.transportRequestId}'"]
        if(backendType == BackendType.SOLMAN)
            uploadingMessage << " of change document '${configuration.changeDocumentId}'"
        uploadingMessage << '.'

        echo uploadingMessage.join()

            try {


                cm.uploadFileToTransportRequest(backendType,
                                                configuration.changeDocumentId,
                                                configuration.transportRequestId,
                                                configuration.applicationId,
                                                configuration.filePath,
                                                configuration.changeManagement.endpoint,
                                                configuration.changeManagement.credentialsId,
                                                configuration.changeManagement.clientOpts)

            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }


        def uploadedMessage = ["[INFO] File '${configuration.filePath}' has been successfully uploaded to transport request '${configuration.transportRequestId}'"]
        if(backendType == BackendType.SOLMAN)
            uploadedMessage << " of change document '${configuration.changeDocumentId}'"
        uploadedMessage << '.'
        echo uploadedMessage.join()
    }
}
