import groovy.io.FileType

import static groovy.json.JsonOutput.toJson

COMMIT_HASH = null
RUNNING_LOCALLY = false
AUXILIARY_SLEEP_MS = 10000
// Build is killed at 50 min, print log to console at minute 45
PRINT_LOGS_AFTER_45_MINUTES_COUNTDOWN = (45 * 60 * 1000) / AUXILIARY_SLEEP_MS
WORKSPACES_ROOT = 'workspaces'
TEST_CASES_DIR = 'testCases'
LIBRARY_VERSION_UNDER_TEST = null

EXCLUDED_FROM_CONSUMER_TESTING_REGEXES = [
    /^documentation\/.*$/,
    /^.travis.yml$/,
    /^test\/.*$/
]


if (!System.getenv('CX_INFRA_IT_CF_USERNAME') || !System.getenv('CX_INFRA_IT_CF_PASSWORD')) {
    throw new RuntimeException('Environment variables CX_INFRA_IT_CF_USERNAME and CX_INFRA_IT_CF_PASSWORD need to be set.')
}

newEmptyDir(WORKSPACES_ROOT)
TestRunnerThread.workspacesRootDir = WORKSPACES_ROOT
LIBRARY_VERSION_UNDER_TEST = "git log --format=%H -n 1".execute().text.trim()
TestRunnerThread.libraryVersionUnderTest = LIBRARY_VERSION_UNDER_TEST
TestRunnerThread.repositoryUnderTest = System.getenv('TRAVIS_REPO_SLUG') ?: 'SAP/jenkins-library'

def testCaseThreads
def cli = new CliBuilder(
    usage: 'groovy consumerTestController.groovy [<options>]',
    header: 'Options:',
    footer: 'If no options are set all tests are run centrally.')

cli.with {
    h longOpt: 'help', 'Print this help text and exit.'
    l longOpt: 'run-locally', 'Run consumer tests locally.'
    s(longOpt: 'single-test', args: 1, argName: 'filePath', 'Run single test.')
}

def options = cli.parse(args)

if (options.h) {
    cli.usage()
    System.exit 0
}

if (options.l) {
    RUNNING_LOCALLY = true
}

if (!RUNNING_LOCALLY) {
    if (changeDoesNotNeedConsumerTesting()) {
        notifyGithub("success", "No consumer tests necessary.")
        println 'No consumer tests necessary.'
        System.exit(0)
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
    COMMIT_HASH = System.getenv('TRAVIS_PULL_REQUEST_SHA') ?: System.getenv('TRAVIS_COMMIT')

    notifyGithub("pending", "Consumer tests are in progress.")
}

if (options.s) {
    testCaseThreads = [new TestRunnerThread(options.s)]
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
    def singleTestCase = testCaseThreads.size() == 1
    for (; ;) {
        if (singleTestCase) {
            testCaseThreads[0].printRunningStdOut()
        } else {
            println "[INFO] Consumer tests are still running."
        }

        if (!singleTestCase && PRINT_LOGS_AFTER_45_MINUTES_COUNTDOWN-- == 0) {
            testCaseThreads.each { thread ->
                thread.printOutput()
            }
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
    status = "failure"
    statusMessage "The following consumer test(s) failed: ${failedThreads}"
    failedThreads.each { failedThread ->
        println "[ERROR] ${failedThread.uniqueName}: Process execution of command: '${failedThread.lastCommand}' failed. " +
            "Return code: ${failedThread.returnCode}."
        failedThread.printOutput()
    }
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
        threads << new TestRunnerThread(file.toString())
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
        target_url : System.getenv('TRAVIS_BUILD_WEB_URL'),
        description: description,
        context    : "integration-tests"
    ]

    con.setDoOutput(true)
    con.getOutputStream().withStream { os ->
        os.write(toJson(postBody).getBytes("UTF-8"))
    }

    int responseCode = con.getResponseCode()
    if (responseCode != HttpURLConnection.HTTP_CREATED) {
        exitPrematurely(34, // Error code taken from curl: CURLE_HTTP_POST_ERROR
            "[ERROR] Posting status to github failed. Expected response code " +
                "'${HttpURLConnection.HTTP_CREATED}', but got '${responseCode}'. " +
                "Response message: '${con.getResponseMessage()}'")
    }
}

def changeDoesNotNeedConsumerTesting() {
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
            throw new RuntimeException("Deletion of dir '${dirName}' failed.")
        }
    }
    if (!dir.mkdirs()) {
        throw new RuntimeException("Creation of dir '${dirName}' failed.")
    }
}
