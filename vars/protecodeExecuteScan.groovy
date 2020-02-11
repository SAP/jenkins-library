import com.sap.piper.JenkinsUtils
import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/protecode.yaml'

/**
 * Protecode is an Open Source Vulnerability Scanner that is capable of scanning binaries. It can be used to scan docker images but is supports many other programming languages especially those of the C family. You can find more details on its capabilities in the [OS3 - Open Source Software Security JAM](https://jam4.sapjam.com/groups/XgeUs0CXItfeWyuI4k7lM3/overview_page/aoAsA0k4TbezGFyOkhsXFs). For getting access to Protecode please visit the [guide](https://go.sap.corp/protecode).
 */
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,  failOnError: true) {

        def script = checkScript(this, parameters) ?: this

        Map config
        def utils = parameters.juStabUtils ?: new Utils()
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()

        new PiperGoUtils(this, utils).unstashPiperBin()
        utils.unstash('pipelineConfigAndTests')

        writeFile(file: METADATA_FILE, text: libraryResource(METADATA_FILE))

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(parameters)}",
        ]) {

            // get context configuration
            config = readJSON (text: sh(returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '${METADATA_FILE}'"))

            def creds = []
            if (config.protecodeCredentialsId) creds.add(usernamePassword(credentialsId: config.protecodeCredentialsId, passwordVariable: 'PIPER_password', usernameVariable: 'PIPER_user'))
            if (config.dockerCredentialsId) creds.add(file(credentialsId: config.dockerCredentialsId, variable: 'FILE_PATH'))

            // execute step
            withCredentials(creds) {

                if(config.dockerCredentialsId) {
                    if (FILE_PATH) {
                        File file = new File(FILE_PATH);
                        def configDirPath = file.getAbsoluteFile().getParent();
                        withEnv([
                            "DOCKER_CONFIG=${configDirPath}",
                        ]) {
                            sh "./piper protecodeExecuteScan"
                        }
                    }
                }
                else {
                        sh "./piper protecodeExecuteScan"
                }
            }

            def json = readJSON (file: "protecodescan_vulns.json")

            def report = readJSON (file: 'protecodeExecuteScan.json')

            archiveArtifacts artifacts: report['target'], allowEmptyArchive: !report['mandatory']
            archiveArtifacts artifacts: "protecodeExecuteScan.json", allowEmptyArchive: false
            archiveArtifacts artifacts: "protecodescan_vulns.json", allowEmptyArchive: false
            
            jenkinsUtils.removeJobSideBarLinks("artifact/${report['target']}")
            jenkinsUtils.addJobSideBarLink("artifact/${report['target']}", "Protecode Report", "images/24x24/graph.png")
            jenkinsUtils.addRunSideBarLink("artifact/${report['target']}", "Protecode Report", "images/24x24/graph.png")
            jenkinsUtils.addRunSideBarLink("${report['protecodeServerUrl']}/products/${report['productID']}/", "Protecode WebUI", "images/24x24/graph.png")
        }
    }
}

