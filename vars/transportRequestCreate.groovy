import com.sap.piper.GitUtils
import com.sap.piper.Utils
import groovy.transform.Field

import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException

import static com.sap.piper.cm.StepHelpers.getChangeDocumentId

@Field def STEP_NAME = 'transportRequestCreate'

@Field Set stepConfigurationKeys = [
    'changeManagement',
     'developmentSystemId'
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
            .withMandatoryProperty('changeManagement/clientOpts')
            .withMandatoryProperty('changeManagement/credentialsId')
            .withMandatoryProperty('changeManagement/endpoint')
            .withMandatoryProperty('changeManagement/git/from')
            .withMandatoryProperty('changeManagement/git/to')
            .withMandatoryProperty('changeManagement/git/format')
            .withMandatoryProperty('developmentSystemId')

        Map configuration =  configHelper.use()

        new Utils().pushToSWA([step: STEP_NAME], configuration)

        def changeDocumentId = getChangeDocumentId(cm, this, configuration)

        configuration = configHelper.mixin([changeDocumentId: changeDocumentId?.trim() ?: null], ['changeDocumentId'] as Set)
                                    .withMandatoryProperty('changeDocumentId',
                                        "Change document id not provided (parameter: \'changeDocumentId\' or via commit history).")
                                    .use()

        def transportRequestId

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
        return transportRequestId
    }
}
