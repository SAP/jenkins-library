package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

@NonCPS
def getMandatoryParameter(Map map, paramName, defaultValue) {

    def paramValue = map[paramName]

    if (paramValue == null)
        paramValue = defaultValue

    if (paramValue == null)
        throw new Exception("ERROR - NO VALUE AVAILABLE FOR ${paramName}")
    return paramValue

}

def retrieveGitCoordinates(script){
    def gitUrl = script.params.GIT_URL
    def gitBranch = script.params.GIT_BRANCH
    if(!gitUrl && !gitBranch) {
        echo "[INFO] Parameters 'GIT_URL' and 'GIT_BRANCH' not set in Jenkins job configuration. Assuming application to be built is contained in the same repository as this Jenkinsfile."
        gitUrl = scm.userRemoteConfigs[0].url
        gitBranch = scm.branches[0].name
    }
    else if(!gitBranch) {
        error "Parameter 'GIT_BRANCH' not set in Jenkins job configuration. Either set both GIT_URL and GIT_BRANCH of the application to be built as Jenkins job parameters or put this Jenkinsfile into the same repository as the application to be built."
    }
    else if(!gitUrl) {
        error "Parameter 'GIT_URL' not set in Jenkins job configuration. Either set both GIT_URL and GIT_BRANCH of the application to be built as Jenkins job parameters or put this Jenkinsfile into the same repository as the application to be built."
    }
    echo "[INFO] Building '${gitBranch}@${gitUrl}'."

    return [url: gitUrl, branch: gitBranch]
}

