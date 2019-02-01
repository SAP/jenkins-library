package com.sap.piper

class JenkinsJobController implements Serializable {
    def script
    def timeout = 3600
    String jenkinsUrl = "http://localhost:8080"
    String jobName

    //Trigger scanning of the multi branch builds
    def buildJob() {
        script.sh "curl -s -X POST ${jenkinsUrl}/job/${jobName}/build"
    }

    def getBuildUrl(String branch) {
        return "${jenkinsUrl}/job/${jobName}/job/${URLEncoder.encode(branch, "UTF-8")}/lastBuild/"
    }

    def waitForJobStatus(String branch, String status) {
        def buildUrl = getBuildUrl(branch)
        def timePerLoop = 10

        for (int i = 0; i < timeout; i += timePerLoop) {
            script.sleep timePerLoop
            try {
                script.echo "Checking Build Status of ${jobName} ${branch}"
                def buildInformation = retrieveBuildInformation(branch)

                if (buildInformation.building) {
                    script.echo "Build is still in progress"
                    continue
                }
                if (buildInformation.result == status) {
                    return true
                }
            } catch (Exception e) {
                script.echo "Could not retrieve status for ${buildUrl}. Retrying..."
                script.echo e.getMessage()
                continue
            }
            return false
        }
        script.error("Timeout: Build of waiting for ${jobName} ${branch}")
    }

    def getConsoleText(String branch) {
        def consoleUrl = this.getBuildUrl(branch) + "/consoleText"
        return script.fetchUrl(consoleUrl)
    }

    def printConsoleText(String branch) {
        String consoleOutput = getConsoleText(branch)

        script.echo '***********************************************'
        script.echo '** Begin Output of Example Application Build **'
        script.echo '***********************************************'

        script.echo consoleOutput

        script.echo '*********************************************'
        script.echo '** End Output of Example Application Build **'
        script.echo '*********************************************'
    }

    def retrieveBuildInformation(String branch) {
        def buildUrl = getBuildUrl(branch)
        def url = "${buildUrl}/api/json"
        script.echo "Checking Build Status of ${jobName} ${branch}"
        script.echo "${jenkinsUrl}/job/${jobName}/job/${URLEncoder.encode(branch, "UTF-8")}/"
        def response = script.fetchUrl(url)
        def result = script.readJSON text: response
        return result
    }

    def waitForSuccess(String branch) {
        if (this.waitForJobStatus(branch, 'SUCCESS')) {
            script.echo "Build was successful"
        } else {
            this.printConsoleText(branch)
            script.error("Build of ${jobName} ${branch} was not successfull")
        }
    }

    def waitForError(String branch, String errorMessage) {
        if (!this.waitForJobStatus(branch, 'FAILURE')) {
            script.echo "Build of ${jobName} ${branch} was successfull but should fail"
        } else {
            String consoleOutput = getConsoleText(branch)

            if (!consoleOutput.contains(errorMessage)) {
                this.printConsoleText(branch)
                script.error("Build of ${jobName} ${branch} failed with a different error message than: ${errorMessage}.")
            }

            script.echo "Build of ${jobName} ${branch} failed as expected"
        }
    }
}
