import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.Utils
import groovy.transform.Field

import com.sap.piper.GenerateDocumentation
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
         * AS ABAP instance number. Only for `RFC`.
         * @parentConfigKey changeManagement
         */
        'rfc/developmentInstance',
        /**
         * AS ABAP client number. Only for `RFC`.
         * @parentConfigKey changeManagement
         */
        'rfc/developmentClient',
    /**
    * The logical system id for which the transport request is created.
    * The format is `<SID>~<TYPE>(/<CLIENT>)?`. For ABAP Systems the `developmentSystemId`
    * looks like `DEV~ABAP/100`. For non-ABAP systems the `developmentSystemId` looks like
    * e.g. `L21~EXT_SRV` or `J01~JAVA`. In case the system type is not known (in the examples
    * provided here: `EXT_SRV` or `JAVA`) the information can be retrieved from the Solution Manager instance.
    * Only for `SOLMAN`.
    */
    'developmentSystemId',
    /**
    * The description of the transport request. Only for `CTS`.
    */
    'description',
    /**
    * The system receiving the transport request. Only for `CTS`.
    */
    'targetSystem',
    /**
    * Typically `W` (workbench) or `C` customizing. Only for `CTS`.
    */
    'transportType',
    /**
    * Provides additional details. Only for `RFC`.
    */
    'verbose'
]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /** The id of the change document to that the transport request is bound to.
    * Typically this value is provided via commit message in the commit history.
    * Only for `SOLMAN`.
    */
    'changeDocumentId'
])

/**
* Creates
*
* * a Transport Request for a Change Document on the Solution Manager (type `SOLMAN`) or
* * a Transport Request inside an ABAP system (type`CTS`)
*
* The id of the transport request is available via [commonPipelineEnvironment.getTransportRequestId()](commonPipelineEnvironment.md)
*/
@GenerateDocumentation
void call(Map parameters = [:]) {

    def transportRequestId

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this
        String stageName = parameters.stageName ?: env.STAGE_NAME

        addPipelineWarning(script, "Deprecation Warning", "The step ${STEP_NAME} is deprecated. Follow the documentation for options.")

        ChangeManagement cm = parameters.cmUtils ?: new ChangeManagement(script)

        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .collectValidationFailures()
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
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
            .withMandatoryProperty('changeManagement/rfc/developmentInstance', null, {backendType == BackendType.RFC})
            .withMandatoryProperty('changeManagement/rfc/developmentClient', null, {backendType == BackendType.RFC})
            .withMandatoryProperty('verbose', null, {backendType == BackendType.RFC})

        def changeDocumentId = null

        if(backendType == BackendType.SOLMAN) {

            changeDocumentId = getChangeDocumentId(cm, script, configuration)

            configHelper.mixin([changeDocumentId: changeDocumentId?.trim() ?: null], ['changeDocumentId'] as Set)
                        .withMandatoryProperty('developmentSystemId')
                        .withMandatoryProperty('changeDocumentId',
                            "Change document id not provided (parameter: \'changeDocumentId\' provided to the step call or via commit history).")
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
                        configuration.changeManagement.solman.docker,
                        configuration.changeDocumentId,
                        configuration.developmentSystemId,
                        configuration.changeManagement.endpoint,
                        configuration.changeManagement.credentialsId,
                        configuration.changeManagement.clientOpts
                    )
                } else if(backendType == BackendType.CTS) {
                    transportRequestId = cm.createTransportRequestCTS(
                        configuration.changeManagement.cts.docker,
                        configuration.transportType,
                        configuration.targetSystem,
                        configuration.description,
                        configuration.changeManagement.endpoint,
                        configuration.changeManagement.credentialsId,
                        configuration.changeManagement.clientOpts
                    )
                } else if (backendType == BackendType.RFC) {
                    transportRequestId = cm.createTransportRequestRFC(
                        configuration.changeManagement.rfc.docker,
                        configuration.changeManagement.endpoint,
                        configuration.changeManagement.rfc.developmentInstance,
                        configuration.changeManagement.rfc.developmentClient,
                        configuration.changeManagement.credentialsId,
                        configuration.description,
                        configuration.verbose
                    )
                } else {
                  throw new IllegalArgumentException("Invalid backend type: '${backendType}'.")
                }
            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }


        echo "[INFO] Transport Request '$transportRequestId' has been successfully created."
        script.commonPipelineEnvironment.setValue('transportRequestId', "${transportRequestId}")
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
