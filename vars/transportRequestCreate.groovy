import com.sap.piper.GitUtils
import groovy.transform.Field

import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException


@Field def STEP_NAME = 'transportRequestCreate'

@Field Set parameterKeys = [
    'changeDocumentId',
    'developmentSystemId',
    'cmCredentialsId',
    'cmEndpoint'
  ]

@Field Set stepConfigurationKeys = [
    'cmCredentialsId',
    'cmEndpoint'
  ]

def call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        ChangeManagement cm = new ChangeManagement(script)

        Map configuration = ConfigurationMerger.merge(parameters.script, STEP_NAME,
                                                      parameters, parameterKeys,
                                                      stepConfigurationKeys)

        def changeDocumentId = configuration.changeDocumentId
        if(!changeDocumentId) throw new AbortException('Change document id not provided (parameter: \'changeDocumentId\').')

        def developmentSystemId = configuration.developmentSystemId
        if(!developmentSystemId) throw new AbortException('Development system id not provided (parameter: \'developmentSystemId\').')

        def cmCredentialsId = configuration.cmCredentialsId
        if(!cmCredentialsId) throw new AbortException('Credentials id not provided (parameter: \'cmCredentialsId\').')

        def cmEndpoint = configuration.cmEndpoint
        if(!cmEndpoint) throw new AbortException('Solution Manager endpoint not provided (parameter: \'cmEndpoint\').')

        def transportRequestId

        echo "[INFO] Creating transport request for change document '$changeDocumentId' and development system '$developmentSystemId'."

        withCredentials([usernamePassword(
            credentialsId: cmCredentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {

            try {
                transportRequestId = cm.createTransportRequest(changeDocumentId, developmentSystemId, cmEndpoint, username, password)
            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }
        }

        echo "[INFO] Transport Request '$transportRequestId' has been successfully created."
        return transportRequestId
    }
}
