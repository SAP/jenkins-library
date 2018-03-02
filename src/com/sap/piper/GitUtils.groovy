package com.sap.piper

def getGitCommitId() {
    return sh(returnStdout: true, script: 'git rev-parse HEAD').trim()
}
