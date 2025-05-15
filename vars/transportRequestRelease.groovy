import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.Utils
import groovy.transform.Field

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.cm.BackendType
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException

import static com.sap.piper.cm.StepHelpers.getTransportRequestId
import static com.sap.piper.cm.StepHelpers.getChangeDocumentId
import static com.sap.piper.cm.StepHelpers.getBackendTypeAndLogInfoIfCMIntegrationDisabled

@Field def STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

@Field Set STEP_CONFIG_KEYS = [
    'changeManagement',
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
  ]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /** @see transportRequestCreate */
    'changeDocumentId',
    /** The id of the transport request to release. */
    'transportRequestId',
    /** @see transportRequestCreate */
    'verbose',
])

/** Releases a Transport Request. */
@GenerateDocumentation
void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this
        String stageName = parameters.stageName ?: env.STAGE_NAME

        addPipelineWarning(script, "Deprecation Warning", "The step ${STEP_NAME} is deprecated. Follow the documentation for options.")

        ChangeManagement cm = parameters.cmUtils ?: new ChangeManagement(script)

        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)

        Map configuration = configHelper.use()

        BackendType backendType = getBackendTypeAndLogInfoIfCMIntegrationDisabled(this, configuration)
        if(backendType == BackendType.NONE) return

        configHelper
            .collectValidationFailures()
            .withMandatoryProperty('changeManagement/clientOpts')
            .withMandatoryProperty('changeManagement/credentialsId')
            .withMandatoryProperty('changeManagement/endpoint')
            .withMandatoryProperty('changeManagement/git/to')
            .withMandatoryProperty('changeManagement/git/from')
            .withMandatoryProperty('changeManagement/git/format')
            .withMandatoryProperty('changeManagement/rfc/developmentInstance', null, { backendType == BackendType.RFC})
            .withMandatoryProperty('changeManagement/rfc/developmentClient', null, { backendType == BackendType.RFC})
            .withMandatoryProperty('verbose', null, { backendType == BackendType.RFC})

        configuration = configHelper.use()

        def changeDocumentId = null
        def transportRequestId = getTransportRequestId(cm, script, configuration)

        if(backendType == BackendType.SOLMAN) {

            changeDocumentId = getChangeDocumentId(cm, script, configuration)

            configHelper.mixin([changeDocumentId: changeDocumentId?.trim() ?: null], ['changeDocumentId'] as Set)
                        .withMandatoryProperty('changeDocumentId',
                            "Change document id not provided (parameter: \'changeDocumentId\' provided to the step call or via commit history).")

        }

        configuration = configHelper
                            .mixin([transportRequestId: transportRequestId?.trim() ?: null], ['transportRequestId'] as Set)
                            .withMandatoryProperty('transportRequestId',
                                "Transport request id not provided (parameter: \'transportRequestId\' provided to the step call or via commit history).")
                            .use()

        def closingMessage = ["[INFO] Closing transport request '${configuration.transportRequestId}'"]
        if(backendType == BackendType.SOLMAN) closingMessage << " for change document '${configuration.changeDocumentId}'"
        closingMessage << '.'
        echo closingMessage.join()

            try {

                switch(backendType) {

                    case BackendType.SOLMAN:

                        cm.releaseTransportRequestSOLMAN(
                            configuration.changeManagement.solman.docker,
                            configuration.changeDocumentId,
                            configuration.transportRequestId,
                            configuration.changeManagement.endpoint,
                            configuration.changeManagement.credentialsId,
                            configuration.changeManagement.clientOpts)
                        break

                    case BackendType.CTS:

                        cm.releaseTransportRequestCTS(
                            configuration.changeManagement.cts.docker,
                            configuration.transportRequestId,
                            configuration.changeManagement.endpoint,
                            configuration.changeManagement.credentialsId,
                            configuration.changeManagement.clientOpts)
                        break

                    case BackendType.RFC:

                        cm.releaseTransportRequestRFC(
                            configuration.changeManagement.rfc.docker,
                            configuration.transportRequestId,
                            configuration.changeManagement.endpoint,
                            configuration.changeManagement.rfc.developmentInstance,
                            configuration.changeManagement.rfc.developmentClient,
                            configuration.changeManagement.credentialsId,
                            configuration.verbose)
                        break

                    default:

                        throw new IllegalArgumentException("Invalid backend type: '${backendType}'.")
                }
            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }


        echo "[INFO] Transport Request '${configuration.transportRequestId}' has been successfully closed."
    }
}

static void addPipelineWarning(Script script, String heading, String message) {
    script.echo '[WARNING] ' + message
    script.addBadge(icon: "warning.gif", text: message)

    String html =
        """
            <h2>$heading</h2>
            <p>$message</p>
            """

    script.createSummary(icon: "warning.gif", text: html)
}
