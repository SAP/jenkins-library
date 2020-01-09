import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.Utils
import groovy.transform.Field

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.BackendType
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException

import static com.sap.piper.cm.StepHelpers.getTransportRequestId
import static com.sap.piper.cm.StepHelpers.getChangeDocumentId
import static com.sap.piper.cm.StepHelpers.getBackendTypeAndLogInfoIfCMIntegrationDisabled

@Field def STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    'changeManagement',
        /**
         * @see checkChangeInDevelopment
         * @parentConfigKey changeManagement
         */
        'changeDocumentLabel',
        /**
        * Defines where the transport request is created, e.g. SAP Solution Manager, ABAP System.
        * @parentConfigKey changeManagement
        * @possibleValues `SOLMAN`, `CTS`, `RFC`
        */
        'type',
        /**
         * @see checkChangeInDevelopment
         * @parentConfigKey changeManagement
         */
        'clientOpts',
        /**
         * @see checkChangeInDevelopment
         * @parentConfigKey changeManagement
         */
        'credentialsId',
        /**
         * @see checkChangeInDevelopment
         * @parentConfigKey changeManagement
         */
        'endpoint',
        /**
         * @see checkChangeInDevelopment
         * @parentConfigKey changeManagement
         */
        'git/from',
        /**
         * @see checkChangeInDevelopment
         * @parentConfigKey changeManagement
         */
        'git/to',
        /**
         * @see checkChangeInDevelopment
         * @parentConfigKey changeManagement
         */
        'git/format',
        /**
         * @see transportRequestCreate
         * @parentConfigKey changeManagement
         */
        'rfc/developmentInstance',
        /**
         * @see transportRequestCreate
         * @parentConfigKey changeManagement
         */
        'rfc/developmentClient',
        /**
         * A pattern used for identifying lines holding the transport request id.
         * @parentConfigKey changeManagement
         */
        'changeManagement/transportRequestLabel'
  ]

@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
        'applicationName', // RFC
        /** The id of the application. Only for `SOLMAN`.*/
        'applicationId', // SOLMAN
        'applicationDescription',
        /** The path of the file to upload.*/
        'filePath', // SOLMAN, CTS
        /** The URL where to find the UI5 package to upload to the transport request.  Only for `RFC`. */
        'applicationUrl', // RFC
        /** The ABAP package name of your application. */
        'abapPackage',
        /** The code page of your ABAP system. E.g. UTF-8. */
        'codePage', //RFC
        'acceptUnixStyleLineEndings', // RFC
        /** @see transportRequestCreate */
        'verbose', // RFC
    ])

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /** @see transportRequestCreate */
    'changeDocumentId',
    /** The id of the transport request to upload the file.*/
    'transportRequestId'])

/** Uploads a file to a Transport Request. */
@GenerateDocumentation
void call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this

        ChangeManagement cm = parameters.cmUtils ?: new ChangeManagement(script)

        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this, script)
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
            .collectValidationFailures()
            .withMandatoryProperty('changeManagement/changeDocumentLabel')
            .withMandatoryProperty('changeManagement/clientOpts')
            .withMandatoryProperty('changeManagement/credentialsId')
            .withMandatoryProperty('changeManagement/endpoint')
            .withMandatoryProperty('changeManagement/type')
            .withMandatoryProperty('changeManagement/git/from')
            .withMandatoryProperty('changeManagement/git/to')
            .withMandatoryProperty('changeManagement/git/format')
            .withMandatoryProperty('filePath', null, { backendType in [BackendType.SOLMAN, BackendType.CTS] })
            .withMandatoryProperty('applicationUrl', null, { backendType == BackendType.RFC })
            .withMandatoryProperty('codePage', null, { backendType == BackendType.RFC })
            .withMandatoryProperty('acceptUnixStyleLineEndings', null, { backendType == BackendType.RFC })
            .withMandatoryProperty('changeManagement/rfc/developmentInstance', null, { backendType == BackendType.RFC })
            .withMandatoryProperty('changeManagement/rfc/developmentClient', null, { backendType == BackendType.RFC })
            .withMandatoryProperty('changeManagement/rfc/docker/image', null, {backendType == BackendType.RFC})
            .withMandatoryProperty('changeManagement/rfc/docker/options', null, {backendType == BackendType.RFC})
            .withMandatoryProperty('changeManagement/rfc/docker/envVars', null, {backendType == BackendType.RFC})
            .withMandatoryProperty('changeManagement/rfc/docker/pullImage', null, {backendType == BackendType.RFC})
            .withMandatoryProperty('applicationDescription', null, { backendType == BackendType.RFC })
            .withMandatoryProperty('abapPackage', null, { backendType == BackendType.RFC })
            .withMandatoryProperty('applicationId', null, {backendType == BackendType.SOLMAN})
            .withMandatoryProperty('applicationName', null, {backendType == BackendType.RFC})
            .withMandatoryProperty('failOnWarning', null, {backendType == BackendType.RFC})
            .withMandatoryProperty('verbose', null, {backendType == BackendType.RFC})

        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'changeManagementType',
            stepParam1: configuration.changeManagement.type,
            stepParamKey2: 'scriptMissing',
            stepParam2: parameters?.script == null
        ], configuration)

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
        }
        configuration = configHelper
            .withMandatoryProperty('transportRequestId',
                "Transport request id not provided (parameter: \'transportRequestId\' or via commit history).")
            .use()

        def uploadingMessage = ['[INFO] Uploading file ' +
            "'${backendType == BackendType.RFC ? configuration.applicationUrl : configuration.filePath}' " +
            "to transport request '${configuration.transportRequestId}'"]
        if(backendType == BackendType.SOLMAN)
            uploadingMessage << " of change document '${configuration.changeDocumentId}'"
        uploadingMessage << '.'

        echo uploadingMessage.join()

            try {

                switch(backendType) {

                    case BackendType.SOLMAN:
                        cm.uploadFileToTransportRequestSOLMAN(
                            configuration.changeManagement.solman?.docker ?: [:],
                            configuration.changeDocumentId,
                            configuration.transportRequestId,
                            configuration.applicationId,
                            configuration.filePath,
                            configuration.changeManagement.endpoint,
                            configuration.changeManagement.credentialsId,
                            configuration.changeManagement.clientOpts)
                        break
                    case BackendType.CTS:
                        cm.uploadFileToTransportRequestCTS(
                            configuration.changeManagement.cts?.docker ?: [:],
                            configuration.transportRequestId,
                            configuration.filePath,
                            configuration.changeManagement.endpoint,
                            configuration.changeManagement.credentialsId,
                            configuration.changeManagement.clientOpts)
                        break
                    case BackendType.RFC:

                        cm.uploadFileToTransportRequestRFC(
                            configuration.changeManagement.rfc.docker ?: [:],
                            configuration.transportRequestId,
                            configuration.applicationName,
                            configuration.applicationUrl,
                            configuration.changeManagement.endpoint,
                            configuration.changeManagement.credentialsId,
                            configuration.changeManagement.rfc.developmentInstance,
                            configuration.changeManagement.rfc.developmentClient,
                            configuration.applicationDescription,
                            configuration.abapPackage,
                            configuration.codePage,
                            configuration.acceptUnixStyleLineEndings,
                            configuration.failOnWarning,
                            configuration.verbose
                        )

                        break

                }

            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }


        def uploadedMessage = ["[INFO] File '${backendType == BackendType.RFC ? configuration.applicationUrl : configuration.filePath}' has been successfully uploaded to transport request '${configuration.transportRequestId}'"]
        if(backendType == BackendType.SOLMAN)
            uploadedMessage << " of change document '${configuration.changeDocumentId}'"
        uploadedMessage << '.'
        echo uploadedMessage.join()
    }
}
