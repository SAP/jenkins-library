package com.sap.piper

boolean insideWorkTree() {
    return sh(returnStatus: true, script: 'git rev-parse --is-inside-work-tree 1>/dev/null 2>&1') == 0
}

boolean isMergeCommit() throws MissingPropertyException{
    for (def extension : scm.getExtensions()) {
        if(extension instanceof jenkins.plugins.git.MergeWithGitSCMExtension){
            return true;
        }
    }

    return false;
}

String getMergeCommitSha() throws MissingPropertyException{
    return pullRequest.mergeCommitSha
}

boolean isWorkTreeDirty() {

    if(!insideWorkTree()) error 'Method \'isWorkTreeClean\' called outside a git work tree.'

    def gitCmd = 'git diff --quiet HEAD'
    def rc = sh(returnStatus: true, script: gitCmd)

    // from git man page:
    // "it exits with 1 if there were differences and 0 means no differences"
    //
    // in case of general git trouble, e.g. outside work tree this is indicated by
    // a return code higher than 1.
    if(rc == 0) return false
    else if (rc == 1) return true
    else error "git command '${gitCmd}' return with code '${rc}'. This indicates general trouble with git."
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

String[] extractLogLines(
    String filter = '',
    String from = 'origin/master',
    String to = 'HEAD',
    String format = '%b'
) {

    // Checks below: there was an value provided from outside, but the value was null.
    // Throwing an exception is more transparent than making a fallback to the defaults
    // used in case the parameter is omitted in the signature.
    if(filter == null) throw new IllegalArgumentException('Parameter \'filter\' not provided.')
    if(! from?.trim()) throw new IllegalArgumentException('Parameter \'from\' not provided.')
    if(! to?.trim()) throw new IllegalArgumentException('Parameter \'to\' not provided.')
    if(! format?.trim()) throw new IllegalArgumentException('Parameter \'format\' not provided.')

    def gitLogLines = sh ( returnStdout: true,
        script: """#!/bin/bash
            git log --pretty=format:${format} ${from}..${to}
        """
    )?.split('\n')

    // spread not supported here (CPS)
    if(gitLogLines) {
        def trimmedGitLogLines = []
        for(def gitLogLine : gitLogLines) {
            trimmedGitLogLines << gitLogLine.trim()
        }
        return trimmedGitLogLines.findAll { line -> line ==~ /${filter}/ }
    }
    return new String[0]

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
    steps.stash stashName //TODO: should use new Utils().stash
    // return stash name
    return stashName
}
