import groovy.transform.Field

import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

import hudson.AbortException


@Field def STEP_NAME = 'transportRequestGetOrCreate'

@Field Set parameterKeys = [
    'changeDocumentId',
    'developmentSystemId',
    'credentialsId',
    'endpoint'
  ]

@Field Set stepConfigurationKeys = [
    'credentialsId',
    'endpoint'
  ]

def call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        ChangeManagement cm = new ChangeManagement(script)

        Map configuration = ConfigurationMerger.merge(parameters.script, STEP_NAME,
                                                      parameters, parameterKeys,
                                                      stepConfigurationKeys)

        def changeDocumentId = configuration.changeDocumentId
        if(!changeDocumentId) throw new AbortException('Change id not provided (parameter: \'changeDocumentId\').')

        def developmentSystemId = configuration.developmentSystemId
        if(!developmentSystemId) throw new AbortException('Development system id not provided (parameter: \'developmentSystemId\').')

        def credentialsId = configuration.credentialsId
        if(!credentialsId) throw new AbortException('Credentials id not provided (parameter: \'credentialsId\').')

        def endpoint = configuration.endpoint
        if(!endpoint) throw new AbortException('Solution Manager endpoint not provided (parameter: \'endpoint\').')

        def transportRequests

        echo "[INFO] Getting transport requests for change document '$changeDocumentId'."

        withCredentials([usernamePassword(
            credentialsId: credentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {

            try {
                transportRequests = cm.getTransportRequests(changeDocumentId, endpoint, username, password)
            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }

            if(transportRequests.length>1){
                error "Too many open transport requests $transportRequests for change document '$changeDocumentId'."
            }
            else if(!transportRequests || transportRequests.length == 0 || (transportRequests.length == 1 && !transportRequests[0])){
                echo "[INFO] There is no open transport requests for change document '$changeDocumentId'."
                transportRequestId = cm.createTransportRequest(changeDocumentId, developmentSystemId, endpoint, username, password)
            }
            else {
                transportRequestId = transportRequests[0]
                echo "[INFO] Transport request '$transportRequestId' available for change document '$changeDocumentId'."
            }
        }
        return transportRequestId
    }
}

