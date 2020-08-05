import com.sap.piper.JenkinsUtils
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
        def utils = parameters.juStabUtils ?: new Utils()
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()
        String piperGoPath = parameters.piperGoPath ?: './piper'

        piperExecuteBin.prepareExecution(this, utils, parameters)
        piperExecuteBin.prepareMetadataResource(script, METADATA_FILE)
        Map stepParameters = piperExecuteBin.prepareStepParameters(parameters)

        List credentials = [
            [type: 'usernamePassword', id: 'protecodeCredentialsId', env: ['PIPER_username', 'PIPER_password']],
            [type: 'file', id: 'dockerCredentialsId', env: ['DOCKER_CONFIG']],
        ]

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(stepParameters)}",
            "PIPER_correlationID=${env.BUILD_URL}",
        ]) {
            String customDefaultConfig = piperExecuteBin.getCustomDefaultConfigsArg()
            String customConfigArg = piperExecuteBin.getCustomConfigArg(script)

            echo "PIPER_parametersJSON: ${groovy.json.JsonOutput.toJson(stepParameters)}"

            // get context configuration
            Map config
            piperExecuteBin.handleErrorDetails(STEP_NAME) {
                config = piperExecuteBin.getStepContextConfig(script, piperGoPath, METADATA_FILE, customDefaultConfig, customConfigArg)
                echo "Context Config: ${config}"
            }

            // execute step
            try {
                script.commonPipelineEnvironment.writeToDisk(script)
                piperExecuteBin.credentialWrapper(config, credentials){
                    sh "${piperGoPath} ${STEP_NAME}${customDefaultConfig}${customConfigArg}"
                }
            } finally {
                script.commonPipelineEnvironment.readFromDisk(script)
            }

            //TODO: refactor to use jenkinsUtils.StepResults
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
