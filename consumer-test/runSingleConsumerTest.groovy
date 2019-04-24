import static ConsumerTestUtils.exitPrematurely
import static ConsumerTestUtils.newEmptyDir

def filePath = args?.find()
if (!filePath) {
    exitPrematurely(1, 'Need one argument with YAML file path, ' +
        'e.g. `groovy runSingleConsumerTest.groovy testCases/s4sdk/consumer-test-neo.yml`')
}

AUXILIARY_SLEEP_MS = 1000
WORKSPACES_ROOT = 'workspaces'

newEmptyDir(WORKSPACES_ROOT)
ConsumerTestUtils.workspacesRootDir = WORKSPACES_ROOT
ConsumerTestUtils.libraryVersionUnderTest = "git log --format=%H -n 1".execute().text.trim()
ConsumerTestUtils.repositoryUnderTest = System.getenv('TRAVIS_REPO_SLUG') ?: 'SAP/jenkins-library'


def testCaseThread = new TestRunnerThread(filePath)
testCaseThread.start()
waitForTestCase(testCaseThread)


def waitForTestCase(thread) {
    def auxiliaryThread = Thread.start {
        while (thread.isAlive()) {
            sleep(AUXILIARY_SLEEP_MS)
            thread.printRunningStdOut()
            thread.abortIfSevereErrorOccurred()
        }
    }
    auxiliaryThread.join()
}
