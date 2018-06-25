import com.sap.piper.GitUtils
import groovy.transform.Field
import hudson.AbortException

import com.sap.piper.ConfigurationMerger
import com.sap.piper.cm.ChangeManagement
import com.sap.piper.cm.ChangeManagementException

@Field def STEP_NAME = 'checkChangeInDevelopment'

@Field Set parameterKeys = [
    'cmClientOpts',
    'credentialsId',
    'endpoint',
    'failIfStatusIsNotInDevelopment',
    'git_from',
    'git_to',
    'git_label',
    'git_format'
  ]

@Field Set stepConfigurationKeys = [
    'cmClientOpts',
    'credentialsId',
    'endpoint',
    'failIfStatusIsNotInDevelopment',
    'git_from',
    'git_to',
    'git_label',
    'git_format'
  ]

def call(parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        prepareDefaultValues script: this

        GitUtils gitUtils = parameters?.gitUtils ?: new GitUtils()

        ChangeManagement cm = parameters?.cmUtils ?: new ChangeManagement(parameters.script, gitUtils)

        Map configuration = ConfigurationMerger.merge(parameters.script, STEP_NAME,
                                                      parameters, parameterKeys,
                                                      stepConfigurationKeys)


        def changeId

        try {
            changeId = cm.getChangeDocumentId(
                                              configuration.git_from,
                                              configuration.git_to,
                                              configuration.git_label,
                                              configuration.git_format
                                            )

            if(! changeId?.trim()) {
                throw new ChangeManagementException("ChangeId is null or empty.")
            }
        } catch(ChangeManagementException ex) {
            throw new AbortException(ex.getMessage())
        }

        echo "[INFO] ChangeId retrieved from git commit(s): '${changeId}'. " +
             "Commit range: '${configuration.git_from}..${configuration.git_to}'. " +
             "Searching for label '${configuration.git_label}'."

        boolean isInDevelopment

        echo "[INFO] Checking if change document '$changeId' is in development."

        withCredentials([usernamePassword(
            credentialsId: configuration.credentialsId,
            passwordVariable: 'password',
            usernameVariable: 'username')]) {

            try {
                isInDevelopment = cm.isChangeInDevelopment(changeId, configuration.endpoint, username, password, configuration.cmClientOpts)
            } catch(ChangeManagementException ex) {
                throw new AbortException(ex.getMessage())
            }
        }

        if(isInDevelopment) {
            echo "[INFO] Change '${changeId}' is in status 'in development'."
            return true
        } else {
            if(configuration.failIfStatusIsNotInDevelopment.toBoolean()) {
                throw new AbortException("Change '${changeId}' is not in status 'in development'.")

            } else {
                echo "[WARNING] Change '${changeId}' is not in status 'in development'. Failing the pipeline has been explicitly disabled."
                return false
            }
        }
    }
}
