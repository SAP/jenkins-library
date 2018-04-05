package com.sap.piper

def getGitCommitIdOrNull() {
    if (fileExists('.git')) {
        return sh(returnStdout: true, script: 'git rev-parse HEAD').trim()
    } else {
        return null
    }
}

def getGitCommitId() {
    return sh(returnStdout: true, script: 'git rev-parse HEAD').trim()
}
