@Grab('org.yaml:snakeyaml:1.17')

import org.yaml.snakeyaml.Yaml

class TestRunnerThread extends Thread {

    static def workspacesRootDir
    static def libraryVersionUnderTest
    static def repositoryUnderTest

    Process currentProcess
    final StringBuilder stdOut = new StringBuilder()
    final StringBuilder stdErr = new StringBuilder()
    int lastPrintedStdOutLine = -1
    public def returnCode = -1
    public def lastCommand
    def area
    def testCase
    def uniqueName
    def testCaseRootDir
    def testCaseWorkspace
    def testCaseConfig

    TestRunnerThread(File testCaseFile) {
        // Regex pattern expects a folder structure such as '/rootDir/areaDir/testCase.extension'
        def testCaseMatches = (testCaseFile.toString() =~
            /^[\w\-]+\\/([\w\-]+)\\/([\w\-]+)\..*\u0024/)
        this.area = testCaseMatches[0][1]
        this.testCase = testCaseMatches[0][2]
        if (!area || !testCase) {
            throw new RuntimeException("Expecting file structure '/rootDir/areaDir/testCase.yml' " +
                "but got '${testCaseFile}'.")
        }
        this.uniqueName = "${area}|${testCase}"
        this.testCaseRootDir = new File("${workspacesRootDir}/${area}/${testCase}")
        this.testCaseWorkspace = "${testCaseRootDir}/workspace"
        this.testCaseConfig = new Yaml().load(testCaseFile.text)
    }

    void run() {
        println "[INFO] Test case '${uniqueName}' launched."

        if (testCaseRootDir.exists() || !testCaseRootDir.mkdirs()) {
            throw new RuntimeException("Creation of dir '${testCaseRootDir}' failed.")
        }
        executeShell("git clone -b ${testCaseConfig.referenceAppRepoBranch} ${testCaseConfig.referenceAppRepoUrl} " +
            "${testCaseWorkspace}")
        addJenkinsYmlToWorkspace()
        setLibraryVersionInJenkinsfile()

        //Commit the changed version because artifactSetVersion expects the git repo not to be dirty
        executeShell(["git", "-C", "${testCaseWorkspace}", "commit", "--all",
                        '--author="piper-testing-bot <piper-testing-bot@example.com>"',
                        '--message="Set piper lib version for test"'])

        executeShell("docker run -v /var/run/docker.sock:/var/run/docker.sock " +
            "-v ${System.getenv('PWD')}/${testCaseWorkspace}:/workspace -v /tmp " +
            "-e CASC_JENKINS_CONFIG=/workspace/jenkins.yml -e CX_INFRA_IT_CF_USERNAME " +
            "-e CX_INFRA_IT_CF_PASSWORD -e BRANCH_NAME=${testCase} ppiper/jenkinsfile-runner")

        println "*****[INFO] Test case '${uniqueName}' finished successfully.*****"
        printOutput()
    }

    // Configure path to library-repository under test in Jenkins config
    private void addJenkinsYmlToWorkspace() {
        def sourceFile = 'jenkins.yml'
        def sourceText = new File(sourceFile).text.replaceAll(
            '__REPO_SLUG__', repositoryUnderTest)
        def target = new File("${testCaseWorkspace}/${sourceFile}")
        target.write(sourceText)
    }

    // Force usage of library version under test by setting it in the Jenkinsfile,
    // which is then the first definition and thus has the highest precedence.
    private void setLibraryVersionInJenkinsfile() {
        def jenkinsfile = new File("${testCaseWorkspace}/Jenkinsfile")
        def manipulatedText =
            "@Library(\"piper-library-os@${libraryVersionUnderTest}\") _\n" +
                jenkinsfile.text
        jenkinsfile.write(manipulatedText)
    }

    private void executeShell(command) {
        lastCommand = command
        def startOfCommandString = "Shell command: '${command}'\n"
        stdOut << startOfCommandString
        stdErr << startOfCommandString

        currentProcess = command.execute()
        currentProcess.waitForProcessOutput(stdOut, stdErr)

        returnCode = currentProcess.exitValue()

        currentProcess = null

        if (returnCode > 0) {
            throw new ReturnCodeNotZeroException("Test case: [${uniqueName}]; " +
                "shell command '${command} exited with return code '${returnCode}")
        }
    }

    void printOutput() {
        println "\n[INFO] stdout output from test case ${uniqueName}:"
        stdOut.eachLine { line, i ->
            println "${i} [${uniqueName}] ${line}"
            lastPrintedStdOutLine = i
        }

        println "\n[INFO] stderr output from test case ${uniqueName}:"
        stdErr.eachLine { line, i ->
            println "${i} [${uniqueName}] ${line}"
        }
    }

    public void printRunningStdOut() {
        stdOut.eachLine { line, i ->
            if (i > lastPrintedStdOutLine) {
                println "${i} [${uniqueName}] ${line}"
                lastPrintedStdOutLine = i
            }
        }
    }

    @Override
    public String toString() {
        return uniqueName
    }
}

class ReturnCodeNotZeroException extends Exception {
    ReturnCodeNotZeroException(message) {
        super(message)
    }
}
