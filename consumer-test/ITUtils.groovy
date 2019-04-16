import groovy.io.FileType
@Grab(group = 'org.codehaus.groovy.modules.http-builder', module = 'http-builder', version = '0.7')
import groovyx.net.http.RESTClient

import static groovyx.net.http.ContentType.JSON

class ITUtils {

    static def workspacesRootDir
    static def libraryVersionUnderTest
    static def repositoryUnderTest
    static def commitHash

    static def newEmptyDir(String dirName) {
        def dir = new File(dirName)
        if (dir.exists()) {
            dir.deleteDir()
        }
        dir.mkdirs()
    }

    static def listYamlInDirRecursive(String dirname) {
        def dir = new File(dirname)
        def yamlFiles = []
        dir.eachFileRecurse(FileType.FILES) { file ->
            if (file.getName().endsWith('.yml'))
                yamlFiles << file
        }
        return yamlFiles
    }

    static def notifyGithub(state, description) {
        println "[INFO] Notifying about state '${state}' for commit '${commitHash}'."

        def http = new RESTClient("https://api.github.com/repos/SAP/jenkins-library/statuses/" +
            "${commitHash}")
        http.headers['User-Agent'] = 'groovy-script'
        http.headers['Authorization'] = "token ${System.getenv('INTEGRATION_TEST_VOTING_TOKEN')}"

        def postBody = [
            state      : state,
            target_url : System.getenv('TRAVIS_BUILD_WEB_URL'),
            description: description,
            context    : "integration-tests"
        ]

        http.post(body: postBody, requestContentType: JSON) { response ->
            assert response.statusLine.statusCode == 201
        }
    }
}
