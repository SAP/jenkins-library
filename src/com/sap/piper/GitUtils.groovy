package com.sap.piper

boolean insideWorkTree() {
    return sh(returnStatus: true, script: 'git rev-parse --is-inside-work-tree 1>/dev/null 2>&1') == 0
}

String getGitCommitIdOrNull() {
    if ( insideWorkTree() ) {
        return getGitCommitId()
    } else {
        return null
    }
}

String getGitCommitId() {
    return sh(returnStdout: true, script: 'git rev-parse HEAD').trim()
}
