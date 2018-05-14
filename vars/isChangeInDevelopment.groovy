import com.sap.piper.GitUtils
import groovy.transform.Field

@Field def STEP_NAME = 'isChangeInDevelopment'

def call(parameters = [:]) {


    Set parameterKeys = [
        ]

    Set stepConfigurationKeys = [
        ]

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def changeId = getChangeId(gitUtils, 'HEAD~2')
        echo "CHANGE-ID: ${changeId}"
    }
}

String getChangeId(GitUtils gitUtils, String from = 'origin/master', String to = 'HEAD') {

    def label = 'ChangeDocument\\s?:'

    GitUtils gitUtils = new GitUtils()

    if( ! gitUtils.insideWorkTree() ) {
        throw new hudson.AbortException('Cannot retrieve change status. Not in a git work tree.')
    } else {
        echo '[INFO] Inside a git work tree.'
    }

    def log = gitUtils.extractLogLines(".*${label}.*", from, to)

    def changeIds = log.collect { line -> line.replaceAll(label,'').trim() } .unique()
        changeIds.retainAll { line -> ! line.isEmpty() }

    if( changeIds.size() == 0 ) {
        throw new hudson.AbortException('Cannot retrieve changeId from git commits.')
    } else if (changeIds.size() > 1) {
        throw new hudson.AbortException("Multiple ChangeIds found: ${changeIds}.")
    }

    return changeIds.get(0)
}
