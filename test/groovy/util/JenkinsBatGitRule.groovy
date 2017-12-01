package util

import com.lesfurets.jenkins.unit.BasePipelineTest
import org.junit.rules.TemporaryFolder

class JenkinsBatGitRule extends TemporaryFolder {

    final BasePipelineTest testInstance

    final String workspaceFolder


    JenkinsBatGitRule(BasePipelineTest testInstance, String workspaceFolder) {
        super()
        this.testInstance = testInstance
        this.workspaceFolder = workspaceFolder

    }

    @Override
    void create() throws IOException {
        super.create()
        registerFileHandler()
    }

    void registerFileHandler() {

        this.testInstance.helper.registerAllowedMethod("sh", [String.class], {
            String command ->

                def commandSplit = command.split(" ")

                def commandPath = commandSplit[0]

                if (commandPath.contains("git")) {

                    def gitCommand = commandSplit[1]

                    if (gitCommand == "clone") {
                        def gitUrl = commandSplit[2]

                        if (gitUrl != null) {
                            newFolder("bats", "bin")
                            newFile("bats/bin/bats")
                        }
                    }

                } else if (commandPath.contains("bats")) {

                    def batsFile = new File(getRoot(), commandPath)

                    if (!batsFile.exists()) {
                        throw new Exception("Request file doesn't exists")
                    } else {
                        println("bats file found")
                    }
                } else {
                    println("command not recognized")
                }
        })

    }
}
