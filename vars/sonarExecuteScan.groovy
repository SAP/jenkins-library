import com.sap.piper.JenkinsUtils
import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils
import static com.sap.piper.Prerequisites.checkScript
import groovy.transform.Field
import java.nio.charset.StandardCharsets

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/sonar.yaml'

void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def stepParameters = [:].plus(parameters)

        def script = checkScript(this, parameters) ?: this
        stepParameters.remove('script')

        def utils = parameters.juStabUtils ?: new Utils()
        stepParameters.remove('juStabUtils')

        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()
        stepParameters.remove('jenkinsUtilsStub')

        new PiperGoUtils(this, utils).unstashPiperBin()
        utils.unstash('pipelineConfigAndTests')
        script.commonPipelineEnvironment.writeToDisk(script)

        writeFile(file: ".pipeline/tmp/${METADATA_FILE}", text: libraryResource(METADATA_FILE))

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(stepParameters)}",
        ]) {
            // get context configuration
            Map config = readJSON(text: sh(returnStdout: true, script: "./piper getConfig --contextConfig --stepMetadata '.pipeline/tmp/${METADATA_FILE}'"))
            echo "Config: ${config}"

            Map stepConfig = readJSON(text: sh(returnStdout: true, script: "./piper getConfig --stepMetadata '.pipeline/tmp/${METADATA_FILE}'"))
            echo "StepConfig: ${stepConfig}"

            // determine credentials to load
            List credentials = []
            if (config.sonarTokenCredentialsId)
                credentials.add(string(credentialsId: config.sonarTokenCredentialsId, variable: 'PIPER_token'))
            if(isPullRequest()){
                checkMandatoryParameter(config, "owner")
                checkMandatoryParameter(config, "repository")
                if(config.legacyPRHandling) {
                    checkMandatoryParameter(config, "githubTokenCredentialsId")
                    if (config.githubTokenCredentialsId)
                        credentials.add(string(credentialsId: config.githubTokenCredentialsId, variable: 'PIPER_githubToken'))
                }
            }
            // load certificates into cacerts file
            loadCertificates([customTlsCertificateLinks: stepConfig.customTlsCertificateLinks, verbose: stepConfig.verbose])
            // execute step
            dockerExecute(
                script: script,
                dockerImage: config.dockerImage,
                dockerWorkspace: config.dockerWorkspace,
                dockerOptions: config.dockerOptions,
            ) {
                if(!fileExists('.git')) {		
                    utils.unstash('git')		
                }
                withSonarQubeEnv(stepConfig.instance) {
                    withCredentials(credentials) {
                        sh "./piper ${STEP_NAME}"
                    }
                }
                jenkinsUtils.handleStepResults(STEP_NAME, false, false)
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
