import static ConsumerTestUtils.newEmptyDir
import static ConsumerTestUtils.exitPrematurely

class TestRunnerThread extends Thread {

    Process currentProcess
    StringBuilder stdOut
    StringBuilder stdErr
    int lastPrintedStdOutLine
    public def exitCode = 0
    def area
    def testCase
    def uniqueId
    def testCaseRootDir
    def testCaseWorkspace

    TestRunnerThread(testCaseFilePath) {
        this.stdOut = new StringBuilder()
        this.stdErr = new StringBuilder()

        // Regex pattern expects a folder structure such as '/rootDir/areaDir/testCase.extension'
        def testCaseMatches = (testCaseFilePath.toString() =~
            /^[\w\-]+\\/([\w\-]+)\\/([\w\-]+)\..*\u0024/)
        this.area = testCaseMatches[0][1]
        this.testCase = testCaseMatches[0][2]
        if (!area || !testCase) {
            exitPrematurely(2, "Expecting file structure '/rootDir/areaDir/testCase.yml' " +
                "but got '${testCaseFilePath.toString()}'.")
        }
        this.uniqueId = "${area}:${testCase}"
        this.testCaseRootDir = "${ConsumerTestUtils.workspacesRootDir}/${area}/${testCase}"
        this.testCaseWorkspace = "${testCaseRootDir}/workspace"
    }

    void run() {
        println "[INFO] Test case '${uniqueId}' launched."

        newEmptyDir(testCaseRootDir)
        executeShell("git clone -b ${testCase} https://github.com/sap/cloud-s4-sdk-book " +
            "${testCaseWorkspace}")
        addJenkinsYmlToWorkspace()
        manipulateJenkinsfile()

        //Commit the changed version because artifactSetVersion expects the git repo not to be dirty
        executeShell(["git", "-C", "${testCaseWorkspace}", "commit", "--all",
                      "--author=piper-testing-bot <piper-testing-bot@example.com>",
                      "--message=Set piper lib version for test"])

        executeShell("docker run -v /var/run/docker.sock:/var/run/docker.sock " +
            "-v ${System.getenv('PWD')}/${testCaseWorkspace}:/workspace -v /tmp -e " +
            "CASC_JENKINS_CONFIG=/workspace/jenkins.yml -e CX_INFRA_IT_CF_USERNAME -e " +
            "CX_INFRA_IT_CF_PASSWORD -e BRANCH_NAME=${testCase} ppiper/jenkinsfile-runner")

        println "*****[INFO] Test case '${uniqueId}' finished successfully.*****"
        printStdOut()
    }

    // Configure path to library-repository under test in Jenkins config
    private void addJenkinsYmlToWorkspace() {
        def sourceFile = 'jenkins.yml'
        def sourceText = new File(sourceFile).text.replaceAll(
            '__REPO_SLUG__', ConsumerTestUtils.repositoryUnderTest)
        def target = new File("${testCaseWorkspace}/${sourceFile}")
        target.write(sourceText)
    }

    // Force usage of library version under test by setting it in the Jenkinsfile,
    // which is then the first definition and thus has the highest precedence.
    private void manipulateJenkinsfile() {
        def jenkinsfile = new File("${testCaseWorkspace}/Jenkinsfile")
        def manipulatedText =
            "@Library(\"piper-library-os@${ConsumerTestUtils.libraryVersionUnderTest}\") _\n" +
                jenkinsfile.text
        jenkinsfile.write(manipulatedText)
    }

    private void executeShell(command) {
        def startOfCommandString = "Shell command: '${command}'\n"
        stdOut << startOfCommandString
        stdErr << startOfCommandString

        currentProcess = command.execute()
        currentProcess.waitForProcessOutput(stdOut, stdErr)

        exitCode = currentProcess.exitValue()

        def endOfCommandString = "*****Command execution finished with exit code ${exitCode}" +
            ".*****\n\n"
        stdOut << endOfCommandString
        stdErr << endOfCommandString

        currentProcess = null

        if (this.exitCode > 0) {
            synchronized (this) {
                try {
                    wait() // for other threads to print their log first
                    // then it is interrupted
                } catch (InterruptedException e) {
                    printStdOut()
                    printStdErr()
                }
            }
        }
    }

    void printOutputPrematurely() {
        if (this.currentProcess) {
            printStdOut()
            printStdErr()
        } else {
            println "[${testCase}] Warning: Currently no process is running."
        }
    }

    private void printStdOut() {
        if (stdOut) {
            println "\n[INFO] Standard output from test case ${uniqueId}:"
            stdOut.eachLine { line, i ->
                println "${i} [${uniqueId}] ${line}"
                lastPrintedStdOutLine = i
            }
        } else {
            println "\n[WARNING] No standard output for ${uniqueId} exists."
        }
    }

    private void printStdErr() {
        if (stdErr) {
            println "\n[ERROR] Error output from test case ${uniqueId}:"
            stdErr.eachLine { line, i ->
                println "${i} [${uniqueId}] ${line}"
            }
        } else {
            println "\n[WARNING] No error output for ${uniqueId} exists."
        }
    }

    public void printRunningStdOut() {
        abortIfSevereErrorOccurred()
        stdOut?.eachLine { line, i ->
            if (i > lastPrintedStdOutLine) {
                println "${i} [${uniqueId}] ${line}"
                lastPrintedStdOutLine = i
            }
        }
    }

    private void abortIfSevereErrorOccurred() {
        if (stdErr?.find("SEVERE")) {
            printOutputPrematurely()
            exitPrematurely(1, "SEVERE Error in test case ${uniqueId}, aborted!")
        }
    }
}
