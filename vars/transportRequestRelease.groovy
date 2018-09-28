import com.sap.piper.GitUtils
import com.sap.piper.Utils
import groovy.transform.Field

import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException

import static com.sap.piper.cm.StepHelpers.getTransportRequestId
import static com.sap.piper.cm.StepHelpers.getChangeDocumentId

@Field def STEP_NAME = 'transportRequestRelease'

@Field Set stepConfigurationKeys = [
    'changeManagement'
  ]

@Field Set parameterKeys = stepConfigurationKeys.plus([
    'changeDocumentId',
    'transportRequestId',
  ])

@Field Set generalConfigurationKeys = stepConfigurationKeys

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
            .withMandatoryProperty('changeManagement/clientOpts')
            .withMandatoryProperty('changeManagement/credentialsId')
            .withMandatoryProperty('changeManagement/endpoint')
            .withMandatoryProperty('changeManagement/git/to')
            .withMandatoryProperty('changeManagement/git/from')
            .withMandatoryProperty('changeManagement/git/format')

        Map configuration = configHelper.use()

        new Utils().pushToSWA([step: STEP_NAME], configuration)

        def transportRequestId = getTransportRequestId(cm, this, configuration)
        def changeDocumentId = getChangeDocumentId(cm, this, configuration)

        configuration = configHelper
                            .mixin([transportRequestId: transportRequestId?.trim() ?: null,
                                    changeDocumentId: changeDocumentId?.trim() ?: null], ['transportRequestId', 'changeDocumentId'] as Set)
                            .withMandatoryProperty('transportRequestId',
                                "Transport request id not provided (parameter: \'transportRequestId\' or via commit history).")
                            .withMandatoryProperty('changeDocumentId',
                                "Change document id not provided (parameter: \'changeDocumentId\' or via commit history).")
                            .use()

        echo "[INFO] Closing transport request '${configuration.transportRequestId}' for change document '${configuration.changeDocumentId}'."

            try {
                cm.releaseTransportRequest(configuration.changeDocumentId,
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
