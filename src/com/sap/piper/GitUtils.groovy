package com.sap.piper

import com.cloudbees.groovy.cps.NonCPS

def getGitCommitId() {
    return sh(returnStdout: true, script: 'git rev-parse HEAD').trim()
}
