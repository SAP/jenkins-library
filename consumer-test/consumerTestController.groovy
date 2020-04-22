import groovy.io.FileType

import static groovy.json.JsonOutput.toJson

COMMIT_HASH = null
RUNNING_LOCALLY = false
AUXILIARY_SLEEP_MS = 10000
START_TIME_MS = System.currentTimeMillis()
WORKSPACES_ROOT = 'workspaces'
TEST_CASES_DIR = 'testCases'
LIBRARY_VERSION_UNDER_TEST = "git log --format=%H -n 1".execute().text.trim()
REPOSITORY_UNDER_TEST = System.getenv('REPOSITORY_UNDER_TEST') ?: System.getenv('TRAVIS_REPO_SLUG') ?: 'SAP/jenkins-library'
BRANCH_NAME = System.getenv('TRAVIS_BRANCH') ?: System.getenv('BRANCH_NAME')

EXCLUDED_FROM_CONSUMER_TESTING_REGEXES = [
    /^documentation\/.*$/,
    /^.travis.yml$/,
    /^test\/.*$/
]

println "Running tests for repository: ${REPOSITORY_UNDER_TEST}, branch: ${BRANCH_NAME}, commit: ${LIBRARY_VERSION_UNDER_TEST}"

newEmptyDir(WORKSPACES_ROOT)
TestRunnerThread.workspacesRootDir = WORKSPACES_ROOT
TestRunnerThread.libraryVersionUnderTest = LIBRARY_VERSION_UNDER_TEST
TestRunnerThread.repositoryUnderTest = REPOSITORY_UNDER_TEST

def testCaseThreads
def cli = new CliBuilder(
    usage: 'groovy consumerTestController.groovy [<options>]',
    header: 'Options:',
    footer: 'If no options are set, all tests are run centrally, i.e. on travisCI.')

cli.with {
    h longOpt: 'help', 'Print this help text and exit.'
    l longOpt: 'run-locally', 'Run consumer tests locally in Docker, i.e. skip reporting of GitHub status.'
    s longOpt: 'single-test', args: 1, argName: 'filePath', 'Run single test.'
}

def options = cli.parse(args)

if (options.h) {
    cli.usage()
    return
}

if (options.l) {
    RUNNING_LOCALLY = true
}

if (!RUNNING_LOCALLY) {
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
    COMMIT_HASH = System.getenv('TRAVIS_PULL_REQUEST_SHA') ?: System.getenv('TRAVIS_COMMIT') ?: LIBRARY_VERSION_UNDER_TEST

    if (changeDoesNotNeedConsumerTesting()) {
        println 'No consumer tests necessary.'
        notifyGithub("success", "No consumer tests necessary.")
        return
    } else {
        notifyGithub("pending", "Consumer tests are in progress.")
    }
}

if (options.s) {
    def file = new File(options.s)
    if (!file.exists()) {
        exitPrematurely("Test case configuration file '${file}' does not exist. " +
            "Please provide path to a configuration file of structure '/rootDir/areaDir/testCase.yml'.")
    }
    testCaseThreads = [new TestRunnerThread(file)]
} else {
    testCaseThreads = listTestCaseThreads()
}

testCaseThreads.each { it ->
    it.start()
}

//The thread below will print to console while the test cases are running.
//Otherwise the job would be canceled after 10 minutes without output.
def done = false
Thread.start {
    def outputWasPrintedPrematurely = false
    def singleTestCase = (testCaseThreads.size() == 1)
    if (singleTestCase) {
        AUXILIARY_SLEEP_MS = 1000 //for a single test case we print the running output every second
    }
    for (; ;) {
        if (singleTestCase) {
            testCaseThreads[0].printRunningStdOut()
        } else {
            println "[INFO] Consumer tests are still running."
        }

        // Build is killed at 50 min, print log to console at minute 45
        int MINUTES_SINCE_START = (System.currentTimeMillis() - START_TIME_MS) / (1000 * 60)
        if (!singleTestCase && MINUTES_SINCE_START > 44 && !outputWasPrintedPrematurely) {
            testCaseThreads.each { thread ->
                thread.printOutput()
            }
            outputWasPrintedPrematurely = true
        }

        sleep(AUXILIARY_SLEEP_MS)
        if (done) {
            break
        }
    }
}

testCaseThreads.each { it ->
    it.join()
}
done = true

def failedThreads = testCaseThreads.findAll { thread ->
    thread.returnCode != 0
}

def status
def statusMessage
if (failedThreads.size() == 0) {
    status = "success"
    statusMessage = "All consumer tests finished successfully. Congratulations!"
} else {
    failedThreads.each { failedThread ->
        println "[ERROR] ${failedThread.uniqueName}: Process execution of command: '${failedThread.lastCommand}' failed. " +
            "Return code: ${failedThread.returnCode}."
        failedThread.printOutput()
    }
    status = "failure"
    statusMessage = "The following consumer test(s) failed: ${failedThreads}"
}

if (!RUNNING_LOCALLY) {
    notifyGithub(status, statusMessage)
}

println statusMessage

if (status == "failure") {
    System.exit(1)
}


def listTestCaseThreads() {
    //Each dir that includes a yml file is a test case
    def threads = []
    new File(TEST_CASES_DIR).traverse(type: FileType.FILES, nameFilter: ~/^.+\.yml\u0024/) { file ->
        threads << new TestRunnerThread(file)
    }
    return threads
}

def notifyGithub(state, description) {
    println "[INFO] Notifying about state '${state}' for commit '${COMMIT_HASH}'."

    URL url = new URL("https://api.github.com/repos/SAP/jenkins-library/statuses/${COMMIT_HASH}")
    HttpURLConnection con = (HttpURLConnection) url.openConnection()
    con.setRequestMethod('POST')
    con.setRequestProperty("Content-Type", "application/json; utf-8");
    con.setRequestProperty('User-Agent', 'groovy-script')
    con.setRequestProperty('Authorization', "token ${System.getenv('INTEGRATION_TEST_VOTING_TOKEN')}")

    def postBody = [
        state      : state,
        target_url : System.getenv('TRAVIS_BUILD_WEB_URL') ?: System.getenv('BUILD_WEB_URL'),
        description: description,
        context    : "integration-tests"
    ]

    con.setDoOutput(true)
    con.getOutputStream().withStream { os ->
        os.write(toJson(postBody).getBytes("UTF-8"))
    }

    int responseCode = con.getResponseCode()
    if (responseCode != HttpURLConnection.HTTP_CREATED) {
        exitPrematurely("[ERROR] Posting status to github failed. Expected response code " +
            "'${HttpURLConnection.HTTP_CREATED}', but got '${responseCode}'. " +
            "Response message: '${con.getResponseMessage()}'",
            34) // Error code taken from curl: CURLE_HTTP_POST_ERROR
    }
}

def changeDoesNotNeedConsumerTesting() {
    if (BRANCH_NAME == 'master') {
        return false
    }

    def excludesRegex = '(' + EXCLUDED_FROM_CONSUMER_TESTING_REGEXES.join('|') + ')'

    "git remote add sap https://github.com/SAP/jenkins-library.git".execute().waitFor()
    "git fetch sap".execute().waitFor()
    def diff = "git diff --name-only sap/master ${LIBRARY_VERSION_UNDER_TEST}".execute().text.trim()

    for (def line : diff.readLines()) {
        if (!(line ==~ excludesRegex)) {
            return false
        }
    }

    return true
}

static def newEmptyDir(String dirName) {
    def dir = new File(dirName)
    if (dir.exists()) {
        if (!dir.deleteDir()) {
            exitPrematurely("Deletion of dir '${dirName}' failed.")
        }
    }
    if (!dir.mkdirs()) {
        exitPrematurely("Creation of dir '${dirName}' failed.")
    }
}

static def exitPrematurely(String message, int returnCode = 1) {
    println message
    System.exit(returnCode)
}
