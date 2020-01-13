import com.sap.piper.GenerateDocumentation
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Deprecate
import com.sap.piper.internal.DockerUtils
import com.sap.piper.internal.integration.Protecode
import com.sap.piper.internal.JenkinsUtils
import com.sap.piper.internal.Notify
import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils

import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

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

        script.globalPipelineEnvironment.setInfluxStepData('protecode', false)

        new PiperGoUtils(this, utils).unstashPiperBin()
        utils.unstash('pipelineConfigAndTests')

        writeFile(file: METADATA_FILE, text: libraryResource(METADATA_FILE))

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(parameters)}",
        ]) {

        // get context configuration
        config = readJSON (text: sh(returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '${METADATA_FILE}'"))

        if (!config.dockerImage) {
            config = new ConfigurationHelper(config)
                .mixin([
                    dockerImage: script.globalPipelineEnvironment.getAppContainerProperty('dockerMetadata')?.imageNameTag?:script.globalPipelineEnvironment.getDockerMetadata().imageNameTag,
                    dockerRegistryUrl: "${config.dockerRegistryProtocol}://${script.globalPipelineEnvironment.getAppContainerDockerMetadata()?.repo?:script.globalPipelineEnvironment.getDockerMetadata().repo}"
                ])
                .use()
        }

        /* config = new ConfigurationHelper(config)
            .mixin([
                protecodeExcludeCVEs: config.protecodeExcludeCVEs instanceof List?config.protecodeExcludeCVEs:config.protecodeExcludeCVEs?.tokenize(',')
            ])
            .withMandatoryProperty('protecodeGroup')
            .withMandatoryProperty('protecodeCredentialsId')
            .use()
        */
        if (config.dockerImage && !config.filePath) {
            def dockerImageName = new DockerUtils(script).getNameFromImageUrl(config.dockerImage)
            config.filePath = "${dockerImageName.replace('/', '_')}.tar"
        }

        utils.pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], config)


        //FIX
        //TODO github.com/GoogleContainerTools/container-diff/pkg/util sie Mail vom 10.12 Sven
        protecodeDockerWrapper(config, script) {
            
                withCredentials([[$class: 'UsernamePasswordMultiBinding', credentialsId: config.protecodeCredentialsId, passwordVariable: 'password', usernameVariable: 'user']]) {

                    sh "./piper protecodeExecuteScan  --password ${password} --user ${user}"

                    archiveArtifacts artifacts: "${config.reportFileName}", allowEmptyArchive: false
                    if (config.addSideBarLink) {
                        jenkinsUtils.removeJobSideBarLinks("artifact/${config.reportFileName}")
                        jenkinsUtils.addJobSideBarLink("artifact/${config.reportFileName}", "Protecode Report", "images/24x24/graph.png")
                        jenkinsUtils.addRunSideBarLink("artifact/${config.reportFileName}", "Protecode Report", "images/24x24/graph.png")
                        jenkinsUtils.addRunSideBarLink("${config.protecodeServerUrl}/products/${productId}/", "Protecode WebUI", "images/24x24/graph.png")
                    }
                }

                def jsonResult = readJSON file: "VulnResult.txt"

                //check if result is ok else notify
                if(!jsonResult) {
                        Notify.error(this, "Protecode scan failed, please check the log and protecode backend for more details.")
                }

                script.globalPipelineEnvironment.setInfluxCustomDataMapProperty('protecode_data', 'historical_vulnerabilities', jsonResult.historical_vulnerabilities)
                script.globalPipelineEnvironment.setInfluxCustomDataMapProperty('protecode_data', 'triaged_vulnerabilities', jsonResult.triaged_vulnerabilities)
                script.globalPipelineEnvironment.setInfluxCustomDataMapProperty('protecode_data', 'excluded_vulnerabilities', jsonResult.excluded_vulnerabilities)
                script.globalPipelineEnvironment.setInfluxCustomDataMapProperty('protecode_data', 'minor_vulnerabilities', jsonResult.minor_vulnerabilities)
                script.globalPipelineEnvironment.setInfluxCustomDataMapProperty('protecode_data', 'major_vulnerabilities', jsonResult.major_vulnerabilities)
                script.globalPipelineEnvironment.setInfluxCustomDataMapProperty('protecode_data', 'vulnerabilities', jsonResult.vulnerabilities)


                String fileContents = new File("${config.reportFileName}").getText("UTF-8")
                json = script.readJSON text: fileContents

                if(json.results.summary?.verdict?.short == 'Vulns') {
                    echo "${count} ${json.results.summary?.verdict.detailed} of which ${jsonResult.cvss2GreaterOrEqualSeven} had a CVSS v2 score >= 7.0 and ${jsonResult.cvss3GreaterOrEqualSeven} had a CVSS v3 score >= 7.0.\n${jsonResult.excluded_vulnerabilities} vulnerabilities were excluded via configuration (${config.protecodeExcludeCVEs}) and ${jsonResult.triaged_vulnerabilities} vulnerabilities were triaged via the webUI.\nIn addition ${jsonResult.historical_vulnerabilities} historical vulnerabilities were spotted."
                    if(config.protecodeFailOnSevereVulnerabilities && (jsonResult.cvss2GreaterOrEqualSeven > 0 || jsonResult.cvss3GreaterOrEqualSeven > 0)) {
                        Notify.error(this, "Protecode detected Open Source Software Security vulnerabilities, the project is not compliant. For details see the archived report or the web ui: ${config.protecodeServerUrl}/products/${productId}/")
                    }
                }

                script.globalPipelineEnvironment.setInfluxStepData('protecode', true)
             }
        }
    }
}

private void protecodeDockerWrapper(config, script, Closure body) {
    DockerUtils dockerUtils = new DockerUtils(script)
    if (config.dockerImage && dockerUtils.onKubernetes()) {
        dockerExecuteOnKubernetes(
            script: script,
            containerMap: ['docker.wdf.sap.corp:50000/piper/skopeo': 'skopeo']
        ) {
            container('skopeo') {
                dockerUtils.saveImage(config.filePath, config.dockerImage, config.dockerRegistryUrl, config.dockerCredentialsId)
                body()
            }
        }
    } else if (config.dockerImage) {
        dockerUtils.saveImage(config.filePath, config.dockerImage, config.dockerRegistryUrl, config.dockerCredentialsId)
        body()
    } else {
        body()
    }
}
