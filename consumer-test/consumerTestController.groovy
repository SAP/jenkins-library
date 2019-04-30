import groovy.io.FileType

import static ConsumerTestUtils.exitPrematurely
import static ConsumerTestUtils.newEmptyDir
import static ConsumerTestUtils.notifyGithub

AUXILIARY_SLEEP_MS = 10000
// Build is killed at 50 min, print log to console at minute 45
PRINT_LOGS_AFTER_45_MINUTES_COUNTDOWN = (45 * 60 * 1000) / AUXILIARY_SLEEP_MS
WORKSPACES_ROOT = 'workspaces'
TEST_CASES_DIR = 'testCases'

EXCLUDED_FROM_CONSUMER_TESTING_REGEXES = [
    /^documentation\/.*$/,
    /^.travis.yml$/,
    /^test\/.*$/
]

newEmptyDir(WORKSPACES_ROOT)
ConsumerTestUtils.workspacesRootDir = WORKSPACES_ROOT
ConsumerTestUtils.libraryVersionUnderTest = "git log --format=%H -n 1".execute().text.trim()
ConsumerTestUtils.repositoryUnderTest = System.getenv('TRAVIS_REPO_SLUG') ?: 'SAP/jenkins-library'

def testCaseThreads
def filePath = args?.find()
if (filePath) {
    ConsumerTestUtils.runningLocally = true
    AUXILIARY_SLEEP_MS = 1000

    if (!System.getenv('CX_INFRA_IT_CF_USERNAME') || !System.getenv('CX_INFRA_IT_CF_PASSWORD')) {
        exitPrematurely(1, 'Environment variables CX_INFRA_IT_CF_USERNAME and CX_INFRA_IT_CF_PASSWORD need to be set.')
    }

    testCaseThreads = [new TestRunnerThread(filePath)]

} else {
    ConsumerTestUtils.runningLocally = false

    if (changeDoesNotNeedConsumerTesting()) {
        notifyGithub("success", "No consumer tests necessary.")
        exitPrematurely(0, 'No consumer tests necessary.')
    }

    /*
    In case the build is performed for a pull request TRAVIS_COMMIT is a merge
    commit between the base branch and the PR branch HEAD. That commit is actually built.
    But for notifying about a build status we need the commit which is currently
    the HEAD of the PR branch.

    In case the build is performed for a simple branch (not associated with a PR)
    In this case there is no merge commit between any base branch and HEAD of a PR branch.
    The commit which we need for notifying about a build status is in this case simply
    TRAVIS_COMMIT itself.
    */
    ConsumerTestUtils.commitHash = System.getenv('TRAVIS_PULL_REQUEST_SHA') ?: System.getenv('TRAVIS_COMMIT')

    notifyGithub("pending", "Consumer tests are in progress.")

    testCaseThreads = listTestCaseThreads()
}

testCaseThreads.each { it ->
    it.start()
}

//This method will print to console while the test cases are running
//Otherwise the job will be canceled after 10 minutes without output.
waitForTestCases(testCaseThreads)

notifyGithub("success", "All consumer tests succeeded.")


def static listTestCaseThreads() {
    //Each dir that includes a yml file is a test case
    def threads = []
    new File(TEST_CASES_DIR).traverse(type: FileType.FILES, nameFilter: ~/^.+\.yml\u0024/) { file ->
        threads << new TestRunnerThread(file.toString())
    }
    return threads
}

def waitForTestCases(threadList) {
    threadList.metaClass.anyThreadStillAlive = {
        for (thread in delegate) {
            if (thread.isAlive()) {
                return true
            }
        }
        return false
    }

    def auxiliaryThread = Thread.start {
        while (threadList.anyThreadStillAlive()) {
            printOutputOfThreadsIfOneFailed(threadList)

            sleep(AUXILIARY_SLEEP_MS)
            if (PRINT_LOGS_AFTER_45_MINUTES_COUNTDOWN-- == 0) {
                threadList.each { thread ->
                    thread.printOutput()
                }
            }
            if (ConsumerTestUtils.runningLocally) {
                threadList[0].printRunningStdOut()
                threadList[0].abortIfSevereErrorOccurred()
            } else {
                println "[INFO] Consumer tests are still running."
            }
        }
    }
    auxiliaryThread.join()
}

static def printOutputOfThreadsIfOneFailed(threadList) {
    def failedThread = threadList.find { thread ->
        thread.exitCode > 0
    }
    if (failedThread) {
        threadList.each { thread ->
            if (thread.uniqueName != failedThread.uniqueName) {
                thread.printOutput()
                thread.interrupt()
            }
        }
        synchronized (failedThread) {
            failedThread.interrupt()
        }
    }
}

def changeDoesNotNeedConsumerTesting(){
    def excludesRegex = '(' + EXCLUDED_FROM_CONSUMER_TESTING_REGEXES.join('|') + ')'

    "git remote add sap https://github.com/SAP/jenkins-library.git".execute().waitFor()
    "git fetch sap".execute().waitFor()
    def diff = "git diff --name-only sap/master ${ConsumerTestUtils.libraryVersionUnderTest}".execute().text.trim()

    for (def line : diff.readLines()) {
        if (!(line ==~ excludesRegex)) {
            return false
        }
    }

    return true
}
