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

   // Checks below: there was an value provided from outside, but the value was null.
   // Throwing an exception is more transparent than making a fallback to the defaults
   // used in case the paramter is omitted in the signature.
   if(filter == null) throw new IllegalArgumentException('Parameter \'filter\' not provided.')
   if(! from?.trim()) throw new IllegalArgumentException('Parameter \'from\' not provided.')
   if(! to?.trim()) throw new IllegalArgumentException('Parameter \'to\' not provided.')
   if(! format?.trim()) throw new IllegalArgumentException('Parameter \'format\' not provided.')

    sh ( returnStdout: true,
         script: """#!/bin/bash
                    git log --pretty=format:${format} ${from}..${to}
                 """
       )?.split('\n')
        ?.findAll { line -> line ==~ /${filter}/ }

}

static String handleTestRepository(Script steps, Map config){
    def stashName = "testContent-${UUID.randomUUID()}".toString()
    def options = [url: config.testRepository]
    if (config.gitSshKeyCredentialsId)
        options.put('credentialsId', config.gitSshKeyCredentialsId)
    if (config.gitBranch)
        options.put('branch', config.gitBranch)
    // checkout test repository
    steps.git options
    // stash test content
    steps.stash stashName
    // return stash name
    return stashName
}
