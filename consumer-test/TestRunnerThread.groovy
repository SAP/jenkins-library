import ITUtils

class TestRunnerThread extends Thread {

    public Process currentlyRunningProcess
    def area
    def testCase
    def testCaseRootDir
    def testCaseWorkspace

    public TestRunnerThread(testCaseFilePath) {
        // Regex pattern expects a folder structure such as '/rootDir/areaDir/testCase.extension'
        def testCaseMatches = (testCaseFilePath.toString() =~
            /^[\w\-]+\\/([\w\-]+)\\/([\w\-]+)\..*\u0024/)
        this.area = testCaseMatches[0][1]
        this.testCase = testCaseMatches[0][2]
        this.testCaseRootDir = "${ITUtils.workspacesRootDir}/${area}/${testCase}"
        this.testCaseWorkspace = "${testCaseRootDir}/workspace"
    }

    public void run() {
        println "[INFO] Test case '${testCase}' in area '${area}' launched."

        ITUtils.newEmptyDir(testCaseRootDir)
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

        println "[INFO] Test case '${testCase}' in area '${area}' finished."
    }

    // Configure path to library-repository under test in Jenkins config
    private void addJenkinsYmlToWorkspace() {
        def sourceFile = 'jenkins.yml'
        def sourceText = new File(sourceFile).text.replaceAll(
            '__REPO_SLUG__', ITUtils.repositoryUnderTest)
        def target = new File("${testCaseWorkspace}/${sourceFile}")
        target.write(sourceText)
    }

    // Force usage of library version under test by setting it in the Jenkinsfile,
    // which is then the first definition and thus has the highest precedence.
    private void manipulateJenkinsfile() {
        def jenkinsfile = new File("${testCaseWorkspace}/Jenkinsfile")
        def manipulatedText =
            "@Library(\"piper-library-os@${ITUtils.libraryVersionUnderTest}\") _\n" +
            jenkinsfile.text
        jenkinsfile.write(manipulatedText)
    }

    private void executeShell(command) {
        def stdOut = new StringBuilder(), stdErr = new StringBuilder()
        this.currentlyRunningProcess = command.execute()
        this.currentlyRunningProcess.waitForProcessOutput(stdOut, stdErr)
        int exitCode = this.currentlyRunningProcess.exitValue()
        if (exitCode>0) {
            println "[${testCase}] Shell exited with code ${exitCode}."
            println "[${testCase}] Shell command was: '${command}'"
            println "[${testCase}] Console output: ${stdOut}"
            println "[${testCase}] Console error: '${stdErr}'"
            ITUtils.notifyGithub("failure", "The integration tests failed.")
            System.exit(exitCode)
        }
        this.currentlyRunningProcess = null
    }

    public void printStdOut(){
        if (this.currentlyRunningProcess) {
            def stdOut = new StringBuffer()
            this.currentlyRunningProcess.consumeProcessOutputStream(stdOut)
            println "[${testCase}] Console output: ${stdOut}"
        } else {
            println "[${testCase}] Warning: Currently no process is running."
        }
    }
}
