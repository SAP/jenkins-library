import com.sap.piper.GitUtils
import groovy.transform.Field

import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException


@Field def STEP_NAME = 'transportRequestUploadFile'

@Field Set parameterKeys = [
    'changeId',
    'transportRequestId',
    'applicationId',
    'filePath',
    'cmCredentialsId',
    'cmEndpoint'
  ]

@Field Set generalConfigurationKeys = [
    'cmCredentialsId',
    'cmEndpoint'
  ]

def call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        ChangeManagement cm = new ChangeManagement(script)

        Map configuration = ConfigurationMerger.merge(parameters.script, STEP_NAME,
                                                      parameters, parameterKeys,
                                                      generalConfigurationKeys)

        def changeId = configuration.changeId
        if(!changeId) throw new AbortException("Change id not provided (parameter: 'changeId').")

        def transportRequestId = configuration.transportRequestId
        if(!transportRequestId) throw new AbortException("Transport Request id not provided (parameter: 'transportRequestId').")

        def applicationId = configuration.applicationId
        if(!applicationId) throw new AbortException("Application id not provided (parameter: 'applicationId').")

        def filePath = configuration.filePath
        if(!filePath) throw new AbortException("File path not provided (parameter: 'filePath').")

        def cmCredentialsId = configuration.cmCredentialsId
        if(!cmCredentialsId) throw new AbortException("Credentials id not provided (parameter: 'cmCredentialsId').")

        def cmEndpoint = configuration.cmEndpoint
        if(!cmEndpoint) throw new AbortException("Solution Manager endpoint not provided (parameter: 'cmEndpoint').")

        echo "[INFO] Uploading file '$filePath' to transport request '$transportRequestId' of change document '$changeId'."

        withCredentials([usernamePassword(
            credentialsId: cmCredentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {

            try {
                cm.uploadFileToTransportRequest(changeId, transportRequestId, applicationId, filePath, cmEndpoint, username, password)
            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }
        }

        echo "[INFO] File '$filePath' has been successfully uploaded to transport request '$transportRequestId' of change document '$changeId'."
    }
}
