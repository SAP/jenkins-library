import com.sap.piper.GenerateDocumentation
import com.sap.piper.internal.Deprecate
import com.sap.piper.DockerUtils
import com.sap.piper.JenkinsUtils
import com.sap.piper.internal.Notify
import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/protecode.yaml'
@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Whether to create a side bar link pointing to the report produced by Protecode or not
     * @possibleValues `true`, `false`
     **/
    'addSideBarLink',
    /**
     * Decides which parts are removed from the Protecode backend after the scan
     **/
    'cleanupMode',
    /** The ID to the credential used for downloading the docker image from artifactory to scan it with Protecode */
    'dockerCredentialsId',
    /** The reference to the docker image to scan with Protecode */
    'dockerImage',
    /** The reference to the docker registry to scan with Protecode */
    'dockerRegistryUrl',
    /** The URL to fetch the file to scan with Protecode which must be accessible via public HTTP GET request */
    'fetchUrl',
    /** The path to the file from local workspace to scan with Protecode */
    'filePath',
    /** The URL to the Protecode backend */
    'protecodeServerUrl',
    /** The ID of the credentials used to access the Protecode backend */
    'protecodeCredentialsId',
    /** DEPRECATED: Do use triaging within the Protecode UI instead */
    'protecodeExcludeCVEs',
    /**
     * Whether to fail the job on severe vulnerabilties or not
     * @possibleValues `true`, `false`
     **/
    'protecodeFailOnSevereVulnerabilities',
    /** The Protecode group ID of your team */
    'protecodeGroup',
    /** The timeout to wait for the scan to finish */
    'protecodeTimeoutMinutes',
    /** The file name of the report to be created */
    'reportFileName',
    /**
     * Whether to reuse an existing product instead of creating a new one
     * @possibleValues `true`, `false`
     **/
    'reuseExisting',
    /**
     * Whether to the Protecode backend's callback or poll for results
     * @possibleValues `true`, `false`
     **/
    'useCallback',
    /**
     * Whether to log verbose information or not
     * @possibleValues `true`, `false`
     **/
    'verbose'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

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
@GenerateDocumentation
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

            utils.pushToSWA([
                step: STEP_NAME,
                stepParamKey1: 'scriptMissing',
                stepParam1: parameters?.script == null
            ], config)

            callProtecodeScan(config)

            def json = readJSON (file: "Vulns.json")
            
            if(!json) {
                    Notify.error(this, "Protecode scan failed, please check the log and protecode backend for more details.")
            }

            def report = readJSON (file: 'report.json')

            archiveArtifacts artifacts: report['target'], allowEmptyArchive: !report['mandatory']
            
            //move to link report
            jenkinsUtils.removeJobSideBarLinks("artifact/${report['target']}")
            jenkinsUtils.addJobSideBarLink("artifact/${report['target']}", "Protecode Report", "images/24x24/graph.png")
            jenkinsUtils.addRunSideBarLink("artifact/${report['target']}", "Protecode Report", "images/24x24/graph.png")
            jenkinsUtils.addRunSideBarLink("${report['protecodeServerUrl']}/products/${report['productID']}/", "Protecode WebUI", "images/24x24/graph.png")


            if(json.results.summary?.verdict?.short == 'Vulns') {
                echo "${report['count']} ${json.results.summary?.verdict.detailed} of which ${report['cvss2GreaterOrEqualSeven']} had a CVSS v2 score >= 7.0 and ${report['cvss3GreaterOrEqualSeven']} had a CVSS v3 score >= 7.0.\n${report['excludedVulnerabilities']} vulnerabilities were excluded via configuration (${config.protecodeExcludeCVEs}) and ${report['triagedVulnerabilities']} vulnerabilities were triaged via the webUI.\nIn addition ${report['historicalVulnerabilities']} historical vulnerabilities were spotted."
                if(config.protecodeFailOnSevereVulnerabilities && (report['cvss2GreaterOrEqualSeven'] > 0 || report['cvss3GreaterOrEqualSeven'] > 0)) {
                    Notify.error(this, "Protecode detected Open Source Software Security vulnerabilities, the project is not compliant. For details see the archived report or the web ui: ${report['protecodeServerUrl']}/products/${report['productID']}/")
                }
            }
        }
    }
}

private void callProtecodeScan(config) {

    if (config.dockerCredentialsId) {
        echo "with docker credentials"
        withCredentials([file(credentialsId: config.dockerCredentialsId, variable: 'DOCKER_CONFIG')]) {
            withCredentials([[$class: 'UsernamePasswordMultiBinding', credentialsId: config.protecodeCredentialsId, passwordVariable: 'PIPER_password', usernameVariable: 'PIPER_user']]) {
                sh "./piper protecodeExecuteScan --verbose ${config.verbose}"
            }
        }
    } else {
        withCredentials([[$class: 'UsernamePasswordMultiBinding', credentialsId: config.protecodeCredentialsId, passwordVariable: 'PIPER_password', usernameVariable: 'PIPER_user']]) {
            sh "./piper protecodeExecuteScan --verbose ${config.verbose}"
        }
    }
}