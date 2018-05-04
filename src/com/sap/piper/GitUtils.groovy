package com.sap.piper

String getGitCommitIdOrNull() {
    if (fileExists('.git')) {
        return sh(returnStdout: true, script: 'git rev-parse HEAD').trim()
    } else {
        return null
    }
}

String getGitCommitId() {
    return sh(returnStdout: true, script: 'git rev-parse HEAD').trim()
}
