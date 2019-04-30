import groovy.io.FileType

import static groovy.json.JsonOutput.toJson

class ConsumerTestUtils {

    static def workspacesRootDir
    static def libraryVersionUnderTest
    static def repositoryUnderTest
    static def commitHash
    static def runningLocally

    static def newEmptyDir(String dirName) {
        def dir = new File(dirName)
        if (dir.exists()) {
            dir.deleteDir()
        }
        dir.mkdirs()
    }

    static def notifyGithub(state, description) {
        if (!runningLocally) {
            println "[INFO] Notifying about state '${state}' for commit '${commitHash}'."

            URL url = new URL("https://api.github.com/repos/SAP/jenkins-library/statuses/${commitHash}")
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
    }

    static void exitPrematurely(int exitCode, message) {
        println message
        System.exit(exitCode)
    }
}
