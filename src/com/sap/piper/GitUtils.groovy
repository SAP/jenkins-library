package com.sap.piper

import hudson.AbortException

/**
 * Retrieves the git url and branch either from the job parameters or from the configured SCMs.
 * The git url and branch are stored as a map (gitCoordinates) in the commonPipelineEnvironment.
 * @params script The pipeline script to be able to get the parameters from the job configuration.
 * @return The gitCoordinates
 */
def retrieveGitCoordinates(script) {
    def gitUrl = script.params.GIT_URL
    def gitBranch = script.params.GIT_BRANCH
    if (!gitUrl && !gitBranch) {
        echo "[INFO] Parameters 'GIT_URL' and 'GIT_BRANCH' not set in Jenkins job configuration. Assuming application to be built is contained in the same repository as this Jenkinsfile."
        try {
            //[Q] Why not scm.userRemoteConfigs[0].url? [A] To throw an AbortException for the test case.
            gitUrl = retrieveScm().userRemoteConfigs[0].getUrl()
            gitBranch = retrieveScm().branches[0].getName()
        } catch (AbortException e) {
            error "No Source Code Management setup present. If you define the Pipeline directly in the Jenkins job configuration you have to set up parameters GIT_URL and GIT_BRANCH of the application to be built as Jenkins job parameters."
        }
    } else if (!gitBranch) {
        error "Parameter 'GIT_BRANCH' not set in Jenkins job configuration. Either set both GIT_URL and GIT_BRANCH of the application to be built as Jenkins job parameters or put this Jenkinsfile into the same repository as the application to be built."
    } else if (!gitUrl) {
        error "Parameter 'GIT_URL' not set in Jenkins job configuration. Either set both GIT_URL and GIT_BRANCH of the application to be built as Jenkins job parameters or put this Jenkinsfile into the same repository as the application to be built."
    }
    echo "[INFO] Building '${gitBranch}@${gitUrl}'."

    script.commonPipelineEnvironment.setGitCoordinates([url: gitUrl, branch: gitBranch])

    return [url: gitUrl, branch: gitBranch]
}

/*
 * Do not remove/inline this method. The tests rely on it.
 */
def retrieveScm() {
    return scm
}

