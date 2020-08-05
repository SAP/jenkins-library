import com.sap.piper.JenkinsUtils
import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils
import com.sap.piper.analytics.InfluxData
import static com.sap.piper.Prerequisites.checkScript
import groovy.transform.Field
import java.nio.charset.StandardCharsets

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/sonar.yaml'

void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()
        String piperGoPath = parameters.piperGoPath ?: './piper'

        piperExecuteBin.prepareExecution(this, utils, parameters)
        piperExecuteBin.prepareMetadataResource(script, METADATA_FILE)
        Map stepParameters = piperExecuteBin.prepareStepParameters(parameters)

        List credentialInfo = [
            [type: 'token', id: 'sonarTokenCredentialsId', env: ['PIPER_token']],
            [type: 'token', id: 'githubTokenCredentialsId', env: ['PIPER_githubToken']],
        ]

        script.commonPipelineEnvironment.writeToDisk(script)

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
            // get step configuration to access `instance` & `customTlsCertificateLinks` & `owner` & `repository` & `legacyPRHandling`
            Map stepConfig = readJSON(text: sh(returnStdout: true, script: "${piperGoPath} getConfig --stepMetadata '.pipeline/tmp/${METADATA_FILE}'${customDefaultConfig}${customConfigArg}"))
            echo "Step Config: ${stepConfig}"

            List environment = []
            if(isPullRequest()){
                checkMandatoryParameter(stepConfig, "owner")
                checkMandatoryParameter(stepConfig, "repository")
                if(stepConfig.legacyPRHandling) {
                    checkMandatoryParameter(config, "githubTokenCredentialsId")
                }
                environment.add("PIPER_changeId=${env.CHANGE_ID}")
                environment.add("PIPER_changeBranch=${env.CHANGE_BRANCH}")
                environment.add("PIPER_changeTarget=${env.CHANGE_TARGET }")
            }
            try {
                // load certificates into cacerts file
                loadCertificates(customTlsCertificateLinks: stepConfig.customTlsCertificateLinks, verbose: stepConfig.verbose)
                // execute step
                dockerExecute(
                    script: script,
                    dockerImage: config.dockerImage,
                    dockerWorkspace: config.dockerWorkspace,
                    dockerOptions: config.dockerOptions
                ) {
                    if(!fileExists('.git')) utils.unstash('git')
                    piperExecuteBin.handleErrorDetails(STEP_NAME) {
                        withSonarQubeEnv(stepConfig.instance) {
                            withEnv(environment){
                                try {
                                    piperExecuteBin.credentialWrapper(config, credentialInfo){
                                        sh "${piperGoPath} ${STEP_NAME}${customDefaultConfig}${customConfigArg}"
                                    }
                                } finally {
                                    InfluxData.readFromDisk(script)
                                }
                            }
                        }
                        jenkinsUtils.handleStepResults(STEP_NAME, false, false)
                    }
                }
            } finally {
                def ignore = sh script: 'rm -rf .sonar-scanner .certificates', returnStatus: true
            }
        }
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
            sh "keytool ${keytoolOptions.join(' ')} -alias '${filename}' -file '${certificateFolder}${filename}'"
        }
    }
}
