package com.sap.piper

class JenkinsController implements Serializable {
    def script
    def timeout = 3600
    String jenkinsUrl = "http://localhost:8080"

    def createJob(String gitUri, String jobName) {
        String genericJobTemplate = script.libraryResource "jobTemplate.xml"
        String jobXml = Utils.fillTemplate(genericJobTemplate, [guid: UUID.randomUUID().toString(), gitUri: gitUri])
        return this.createJobFromJobXml(jobXml, jobName)
    }

    def waitForJenkinsStarted() {
        def timeout = 120
        def timePerLoop = 5

        for (int i = 0; i < timeout; i += timePerLoop) {
            script.sleep timePerLoop
            try {
                if (retrieveJenkinsStatus() == 'NORMAL') {
                    return true
                }
            } catch (Exception e) {
                script.echo "Could not retrieve status for Jenkins. Retrying..."
                script.echo e.getMessage()
                continue
            }
            return false
        }
        script.error("Timeout: Build of waiting for Jenkins status}")
    }

    JenkinsJobController createJobFromJobXml(String jobXml, String jobName) {
        def retry = 0
        def jobSuccessfullyCreated = false

        def createJobEndpoint = BashUtils.quoteAndEscape("${jenkinsUrl}/createItem?name=${jobName}")


        jobXml = BashUtils.quoteAndEscape(jobXml)

        while (!jobSuccessfullyCreated && retry < 5) {
            script.dir('resources') {
                script.sh "curl -X POST ${createJobEndpoint} -v --header \"Content-Type: application/xml\" -d ${jobXml}"
                script.sleep 5
                jobSuccessfullyCreated = this.isJobAvailable(jobName)
                retry++
            }
            if (jobSuccessfullyCreated) {
                script.echo("Job has been successfully imported to test server")
                return new JenkinsJobController(script: script, jenkinsUrl: jenkinsUrl, jobName: jobName)
            }
        }
        script.error("Job has not been successfully imported to test server")
    }

    def isJobAvailable(String jobName) {
        def jobStateEndpoint = BashUtils.quoteAndEscape("${jenkinsUrl}/job/${jobName}/api/json")
        def statusCode = script.sh(
            script: "curl -XGET -s -o /dev/null -I -w \"%{http_code}\" ${jobStateEndpoint}",
            returnStdout: true
        )
        return statusCode == "200"
    }

    private retrieveJenkinsStatus() {
        def apiUrl = "${jenkinsUrl}/api/json"
        script.echo "Checking Jenkins Status"
        def response = script.fetchUrl(apiUrl)
        def result = script.readJSON text: response
        return result.mode
    }

}
