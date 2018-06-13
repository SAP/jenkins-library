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

String[] extractLogLines(String filter = '',
                         String from = 'origin/master',
                         String to = 'HEAD',
                         String format = '%b') {

    sh ( returnStdout: true,
         script: """#!/bin/bash
                    git log --pretty=format:${format} ${from}..${to}
                 """
       )?.split('\n')
        ?.findAll { line -> line ==~ /${filter}/ }

}
