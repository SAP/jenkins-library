import static com.sap.piper.Prerequisites.checkScript
import static com.sap.piper.BashUtils.quoteAndEscape as q

import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import com.sap.piper.analytics.InfluxData
import groovy.transform.Field
import java.nio.charset.StandardCharsets

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/sonarExecuteScan.yaml'

void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()
        String piperGoPath = parameters.piperGoPath ?: './piper'

        piperExecuteBin.prepareExecution(script, utils, parameters)
        piperExecuteBin.prepareMetadataResource(script, METADATA_FILE)
        Map stepParameters = piperExecuteBin.prepareStepParameters(parameters)

        List credentialInfo = [
            [type: 'token', id: 'sonarTokenCredentialsId', env: ['PIPER_token']],
            [type: 'token', id: 'githubTokenCredentialsId', env: ['PIPER_githubToken']],
        ]

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(stepParameters)}",
            "PIPER_correlationID=${env.BUILD_URL}",
        ]) {
            String customDefaultConfig = piperExecuteBin.getCustomDefaultConfigsArg()
            String customConfigArg = piperExecuteBin.getCustomConfigArg(script)
            // get context configuration
            Map config
            piperExecuteBin.handleErrorDetails(STEP_NAME) {
                config = piperExecuteBin.getStepContextConfig(script, piperGoPath, METADATA_FILE, customDefaultConfig, customConfigArg)
                echo "Context Config: ${config}"
            }
            // get step configuration to access `instance` & `customTlsCertificateLinks` & `owner` & `repository`
            // & `legacyPRHandling` & `inferBranchName`
            // writePipelineEnv needs to be called here as owner and repository may come from the pipeline environment
            writePipelineEnv(script: script, piperGoPath: piperGoPath)
            Map stepConfig = readJSON(text: sh(returnStdout: true, script: "${piperGoPath} getConfig --stepMetadata ${q('.pipeline/tmp/' + METADATA_FILE)}${customDefaultConfig}${customConfigArg}"))
            echo "Step Config: ${stepConfig}"

            List environment = []
            if (isPullRequest()) {
                checkMandatoryParameter(stepConfig, "owner")
                checkMandatoryParameter(stepConfig, "repository")
                if(stepConfig.legacyPRHandling) {
                    checkMandatoryParameter(config, "githubTokenCredentialsId")
                }
            } else if (!isProductiveBranch(script) && stepConfig.inferBranchName && env.BRANCH_NAME) {
                environment.add("PIPER_branchName=${env.BRANCH_NAME}")
            }
            try {
                // load certificates into cacerts file
                if(script.env.PIPER_ENABLE_LEGACY_CERT_LOADING && Boolean.valueOf(script.env.PIPER_ENABLE_LEGACY_CERT_LOADING)){
                    loadCertificates(customTlsCertificateLinks: stepConfig.customTlsCertificateLinks, verbose: stepConfig.verbose)
                }
                // execute step
                piperExecuteBin.dockerWrapper(script, STEP_NAME, config){
                    if(!fileExists('.git')) utils.unstash('git')
                    piperExecuteBin.handleErrorDetails(STEP_NAME) {
                        writePipelineEnv(script: script, piperGoPath: piperGoPath)
                        withEnv(environment) {
                            influxWrapper(script) {
                                try {
                                    piperExecuteBin.credentialWrapper(config, credentialInfo) {
                                        if (stepConfig.instance) {
                                            withSonarQubeEnv(stepConfig.instance) {
                                                echo "Instance is deprecated - please use serverUrl parameter to set URL to the Sonar backend."
                                                sh "${piperGoPath} ${STEP_NAME}${customDefaultConfig}${customConfigArg}"
                                            }
                                        } else {
                                            sh "${piperGoPath} ${STEP_NAME}${customDefaultConfig}${customConfigArg}"
                                        }
                                    }
                                } finally {
                                    jenkinsUtils.handleStepResults(STEP_NAME, false, false)
                                }
                            }
                        }
                    }
                }
            } finally {
                def ignore = sh script: 'rm -rf .sonar-scanner .certificates', returnStatus: true
            }
        }
    }
}

private void influxWrapper(Script script, body){
    try {
        body()
    } finally {
        InfluxData.readFromDisk(script)
    }
}

private void checkMandatoryParameter(config, key){
    if (!config[key]) {
        throw new IllegalArgumentException( "ERROR - NO VALUE AVAILABLE FOR ${key}")
    }
}

private Boolean isPullRequest(){
    return env.CHANGE_ID
}

private Boolean isProductiveBranch(Script script) {
    def productiveBranch = script.commonPipelineEnvironment?.getStepConfiguration('', '')?.productiveBranch
    return env.BRANCH_NAME == productiveBranch
}

private void loadCertificates(Map config) {
    String certificateFolder = '.certificates/'
    List wgetOptions = [
        "--directory-prefix ${certificateFolder}"
    ]
    List keytoolOptions = [
        '-import',
        '-noprompt',
        '-storepass changeit',
        "-keystore ${certificateFolder}cacerts"
    ]
    if (config.customTlsCertificateLinks){
        if(config.verbose){
            wgetOptions.push('--verbose')
            keytoolOptions.push('-v')
        }else{
            wgetOptions.push('--no-verbose')
        }
        config.customTlsCertificateLinks.each { url ->
            def filename = new File(url).getName()
            filename = URLDecoder.decode(filename, StandardCharsets.UTF_8.name())
            sh "wget ${wgetOptions.join(' ')} ${url}"
            sh "keytool ${keytoolOptions.join(' ')} -alias ${q(filename)} -file '${certificateFolder}${filename}'"
        }
    }
}
