import com.sap.piper.GitUtils
import groovy.transform.Field

import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException


@Field def STEP_NAME = 'transportRequestUploadFile'

@Field Set parameterKeys = [
    'changeDocumentId',
    'transportRequestId',
    'applicationId',
    'filePath',
    'credentialsId',
    'endpoint',
    'gitFrom',
    'gitTo',
    'gitFormat',
    'labelTransportRequestId'
  ]

@Field Set generalConfigurationKeys = [
    'credentialsId',
    'endpoint',
    'gitFrom',
    'gitTo',
    'gitFormat',
    'labelTransportRequestId'
  ]

def call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        ChangeManagement cm = parameters.cmUtils ?: new ChangeManagement(script)

        Map configuration = ConfigurationMerger.merge(parameters.script, STEP_NAME,
                                                      parameters, parameterKeys,
                                                      generalConfigurationKeys)

        def changeDocumentId = configuration.changeDocumentId
        if(!changeDocumentId) throw new AbortException("Change document id not provided (parameter: 'changeDocumentId').")

        def transportRequestId = cm.getTransportRequestId(
                                                          script.commonPipelineEnvironment,
                                                          configuration.transportRequestId,
                                                          configuration.gitLabel,
                                                          configuration.gitFrom,
                                                          configuration.gitTo,
                                                          configuration.gitFormat
                                                         )
        if(!transportRequestId) throw new AbortException("Transport Request id not provided (parameter: 'transportRequestId' or via commit history).")

        def applicationId = configuration.applicationId
        if(!applicationId) throw new AbortException("Application id not provided (parameter: 'applicationId').")

        def filePath = configuration.filePath
        if(!filePath) throw new AbortException("File path not provided (parameter: 'filePath').")

        def credentialsId = configuration.credentialsId
        if(!credentialsId) throw new AbortException("Credentials id not provided (parameter: 'credentialsId').")

        def endpoint = configuration.endpoint
        if(!endpoint) throw new AbortException("Solution Manager endpoint not provided (parameter: 'endpoint').")

        echo "[INFO] Uploading file '$filePath' to transport request '$transportRequestId' of change document '$changeDocumentId'."

        withCredentials([usernamePassword(
            credentialsId: credentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {

            try {
                cm.uploadFileToTransportRequest(changeDocumentId, transportRequestId, applicationId, filePath, endpoint, username, password)
            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }
        }

        echo "[INFO] File '$filePath' has been successfully uploaded to transport request '$transportRequestId' of change document '$changeDocumentId'."
    }
}
