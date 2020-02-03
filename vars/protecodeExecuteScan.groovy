import com.sap.piper.JenkinsUtils
import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/protecode.yaml'

/**
 * Protecode is an Open Source Vulnerability Scanner that is capable of scanning binaries. It can be used to scan docker images but is supports many other programming languages especially those of the C family. You can find more details on its capabilities in the [OS3 - Open Source Software Security JAM](https://jam4.sapjam.com/groups/XgeUs0CXItfeWyuI4k7lM3/overview_page/aoAsA0k4TbezGFyOkhsXFs). For getting access to Protecode please visit the [guide](https://go.sap.corp/protecode).
 *
 * !!! info "New: Using executeProtecodeScan for Docker images on JaaS"
 *     **This step now also works on "Jenkins as a Service (JaaS)"!**<br />
 *     For the JaaS use case where the execution happens in a Kubernetes cluster without access to a Docker daemon [skopeo](https://github.com/containers/skopeo) is now used silently in the background to save a Docker image retrieved from a registry.
 *
 *
 * !!! hint "Auditing findings (Triaging)"
 *     Triaging is now supported by the Protecode backend and also Piper does consider this information during the analysis of the scan results though product versions are not supported by Protecode. Therefore please make sure that the `fileName` you are providing does either contain a stable version or that it does not contain one at all. By ensuring that you are able to triage CVEs globally on the upload file's name without affecting any other artifacts scanned in the same Protecode group and as such triaged vulnerabilities will be considered during the next scan and will not fail the build anymore.
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
            if (config.dockerCredentialsId) creds.add(usernamePassword(credentialsId: config.dockerCredentialsId, passwordVariable: 'PIPER_containerRegistryPassword', usernameVariable: 'PIPER_containerRegistryUser'))

            // execute step
            withCredentials(creds) {
                sh "./piper protecodeExecuteScan"
            }

            def json = readJSON (file: "protecodescan_vulns.json")

            def report = readJSON (file: 'protecodescan_report.json')

            archiveArtifacts artifacts: report['target'], allowEmptyArchive: !report['mandatory']
            archiveArtifacts artifacts: "protecodescan_report.json", allowEmptyArchive: false
            archiveArtifacts artifacts: "protecodescan_vulns.json", allowEmptyArchive: false
            
            jenkinsUtils.removeJobSideBarLinks("artifact/${report['target']}")
            jenkinsUtils.addJobSideBarLink("artifact/${report['target']}", "Protecode Report", "images/24x24/graph.png")
            jenkinsUtils.addRunSideBarLink("artifact/${report['target']}", "Protecode Report", "images/24x24/graph.png")
            jenkinsUtils.addRunSideBarLink("${report['protecodeServerUrl']}/products/${report['productID']}/", "Protecode WebUI", "images/24x24/graph.png")
        }
    }
}

