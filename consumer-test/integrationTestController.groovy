import ITUtils

AUXILIARY_SLEEP_MS = 10000
// Build is killed at 50 min, print log to console at minute 45
TERMINATE_AFTER_45_MINUTES = (45 * 60 * 1000) / AUXILIARY_SLEEP_MS
WORKSPACES_ROOT = 'workspaces'
TEST_CASES_DIR = 'testCases'

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
ITUtils.commitHash = System.getenv('TRAVIS_PULL_REQUEST_SHA') ?: System.getenv('TRAVIS_COMMIT')

ITUtils.notifyGithub("pending", "Integration tests are in progress.")

ITUtils.newEmptyDir(WORKSPACES_ROOT)
ITUtils.workspacesRootDir = WORKSPACES_ROOT
ITUtils.libraryVersionUnderTest = "git log --format=%H -n 1".execute().text.trim()
ITUtils.repositoryUnderTest = System.getenv('TRAVIS_REPO_SLUG') ?: 'SAP/jenkins-library'

def testCaseThreads = listTestCaseThreads()
testCaseThreads.each { it ->
    it.start()
}

//This method will print to console while the test cases are running
//Otherwise the job will be canceled after 10 minutes without output.
waitForTestCases(testCaseThreads)

ITUtils.notifyGithub("success", "The integration tests succeeded.")


def listTestCaseThreads() {
    //Each dir that includes a yml file is a test case
    def testCases = ITUtils.listYamlInDirRecursive(TEST_CASES_DIR)
    def threads = []
    testCases.each { file ->
        threads << new TestRunnerThread(file.toString())
    }
    return threads
}

def waitForTestCases(threadList) {
    threadList.metaClass.anyThreadStillAlive = {
        for (thread in delegate) {
            if(thread.isAlive()) {
                return true
            }
        }
        return false
    }

    def auxiliaryThread = Thread.start {
        while (threadList.anyThreadStillAlive()) {
            printOutputOfThreadsIfOneFailed(threadList)

            println "[INFO] Integration tests are still running."
            sleep(AUXILIARY_SLEEP_MS)
            if (TERMINATE_AFTER_45_MINUTES--==0) {
                threadList.each { thread ->
                    thread.printOutputPrematurely()
                }
            }
        }
    }
    auxiliaryThread.join()
}

static def printOutputOfThreadsIfOneFailed(threadList) {
    def failedThread = threadList.find { thread ->
        thread.exitCode>0
    }
    if (failedThread) {
        threadList.each { thread ->
            if (thread.area!=failedThread.area ||
                thread.testCase!=failedThread.testCase) {
                thread.printOutputPrematurely()
                thread.interrupt()
            }
        }
        synchronized (failedThread) {
            failedThread.notify()
        }
    }
}
