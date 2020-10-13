import com.sap.piper.BashUtils
import com.sap.piper.DebugReport
import com.sap.piper.DefaultValueCache
import com.sap.piper.JenkinsUtils
import com.sap.piper.MapUtils
import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

void call(Map parameters = [:], String stepName, String metadataFile, List credentialInfo, boolean failOnMissingReports = false, boolean failOnMissingLinks = false, boolean failOnError = false) {

    handlePipelineStepErrorsParameters = [stepName: stepName, stepParameters: parameters]
    if (failOnError) {
        handlePipelineStepErrorsParameters.failOnError = true
    }

    handlePipelineStepErrors(handlePipelineStepErrorsParameters) {

        Script script = checkScript(this, parameters) ?: this
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()
        def utils = parameters.juStabUtils ?: new Utils()

        String piperGoPath = parameters.piperGoPath ?: './piper'

        prepareExecution(script, utils, parameters)
        prepareMetadataResource(script, metadataFile)
        Map stepParameters = prepareStepParameters(parameters)

        withEnv([
            "PIPER_parametersJSON=${groovy.json.JsonOutput.toJson(stepParameters)}",
            "PIPER_correlationID=${env.BUILD_URL}",
            //ToDo: check if parameters make it into docker image on JaaS
        ]) {
            String defaultConfigArgs = getCustomDefaultConfigsArg()
            String customConfigArg = getCustomConfigArg(script)

            echo "PIPER_parametersJSON: ${groovy.json.JsonOutput.toJson(stepParameters)}"

            // get context configuration
            Map config
            handleErrorDetails(stepName) {
                config = getStepContextConfig(script, piperGoPath, metadataFile, defaultConfigArgs, customConfigArg)
                echo "Context Config: ${config}"
            }

            // prepare stashes
            // first eliminate empty stashes
            config.stashContent = utils.unstashAll(config.stashContent)
            // then make sure that commonPipelineEnvironment, config, ... is also available when step stashing is active
            if (config.stashContent?.size() > 0) {
                config.stashContent.add('pipelineConfigAndTests')
                config.stashContent.add('piper-bin')
            }

            if (parameters.stashNoDefaultExcludes) {
                // Merge this parameter which is only relevant in Jenkins context
                // (for dockerExecuteOnKubernetes step) and go binary doesn't know about
                config.stashNoDefaultExcludes = parameters.stashNoDefaultExcludes
            }

            dockerWrapper(script, stepName, config) {
                handleErrorDetails(stepName) {
                    script.commonPipelineEnvironment.writeToDisk(script)
                    try {
                        credentialWrapper(config, credentialInfo) {
                            sh "${piperGoPath} ${stepName}${defaultConfigArgs}${customConfigArg}"
                        }
                    } finally {
                        jenkinsUtils.handleStepResults(stepName, failOnMissingReports, failOnMissingLinks)
                        script.commonPipelineEnvironment.readFromDisk(script)
                    }
                }
            }
        }
    }
}

// reused in sonarExecuteScan
static void prepareExecution(Script script, Utils utils, Map parameters = [:]) {
    def piperGoUtils = parameters.piperGoUtils ?: new PiperGoUtils(script, utils)
    piperGoUtils.unstashPiperBin()
    utils.unstash('pipelineConfigAndTests')
}

// reused in sonarExecuteScan
static Map prepareStepParameters(Map parameters) {
    Map stepParameters = [:].plus(parameters)

    stepParameters.remove('script')
    stepParameters.remove('jenkinsUtilsStub')
    stepParameters.remove('piperGoPath')
    stepParameters.remove('juStabUtils')
    stepParameters.remove('piperGoUtils')

    // When converting to JSON and back again, entries which had a 'null' value will now have a value
    // of type 'net.sf.json.JSONNull', for which the Groovy Truth resolves to 'true' in for example if-conditions
    return MapUtils.pruneNulls(stepParameters)
}

// reused in sonarExecuteScan
static void prepareMetadataResource(Script script, String metadataFile) {
    script.writeFile(file: ".pipeline/tmp/${metadataFile}", text: script.libraryResource(metadataFile))
}

// reused in sonarExecuteScan
static Map getStepContextConfig(Script script, String piperGoPath, String metadataFile, String defaultConfigArgs, String customConfigArg) {
    return script.readJSON(text: script.sh(returnStdout: true, script: "${piperGoPath} getConfig --contextConfig --stepMetadata '.pipeline/tmp/${metadataFile}'${defaultConfigArgs}${customConfigArg}"))
}

static String getCustomDefaultConfigs() {
    // The default config files were extracted from merged library
    // resources by setupCommonPipelineEnvironment.groovy into .pipeline/.
    List customDefaults = DefaultValueCache.getInstance().getCustomDefaults()
    for (int i = 0; i < customDefaults.size(); i++) {
        customDefaults[i] = BashUtils.quoteAndEscape(".pipeline/${customDefaults[i]}")
    }
    return customDefaults.join(',')
}

// reused in sonarExecuteScan
static String getCustomDefaultConfigsArg() {
    String customDefaults = getCustomDefaultConfigs()
    if (customDefaults) {
        return " --defaultConfig ${customDefaults} --ignoreCustomDefaults"
    }
    return ''
}

// reused in sonarExecuteScan
static String getCustomConfigArg(def script) {
    if (script?.commonPipelineEnvironment?.configurationFile
        && script.commonPipelineEnvironment.configurationFile != '.pipeline/config.yml'
        && script.commonPipelineEnvironment.configurationFile != '.pipeline/config.yaml') {
        return " --customConfig ${BashUtils.quoteAndEscape(script.commonPipelineEnvironment.configurationFile)}"
    }
    return ''
}

// reused in sonarExecuteScan
void dockerWrapper(script, stepName, config, body) {
    if (config.dockerImage) {
        echo "[INFO] executing pipeline step '${stepName}' with docker image '${config.dockerImage}'"
        Map dockerExecuteParameters = [:].plus(config)
        dockerExecuteParameters.script = script
        dockerExecute(dockerExecuteParameters) {
            body()
        }
    } else {
        body()
    }
}

// reused in sonarExecuteScan
void credentialWrapper(config, List credentialInfo, body) {
    if (config.containsKey('vaultAppRoleTokenCredentialsId') && config.containsKey('vaultAppRoleSecretTokenCredentialsId')) {
        credentialInfo = [[type: 'token', id: 'vaultAppRoleTokenCredentialsId', env: ['PIPER_vaultAppRoleID']],
                            [type: 'token', id: 'vaultAppRoleSecretTokenCredentialsId', env: ['PIPER_vaultAppRoleSecretID']]]
    }
    if (credentialInfo.size() > 0) {
        def creds = []
        def sshCreds = []
        credentialInfo.each { cred ->
            switch(cred.type) {
                case "file":
                    if (config[cred.id]) creds.add(file(credentialsId: config[cred.id], variable: cred.env[0]))
                    break
                case "token":
                    if (config[cred.id]) creds.add(string(credentialsId: config[cred.id], variable: cred.env[0]))
                    break
                case "usernamePassword":
                    if (config[cred.id]) creds.add(usernamePassword(credentialsId: config[cred.id], usernameVariable: cred.env[0], passwordVariable: cred.env[1]))
                    break
                case "ssh":
                    if (config[cred.id]) sshCreds.add(config[cred.id])
                    break
                default:
                    error ("invalid credential type: ${cred.type}")
            }
        }

        if (sshCreds.size() > 0) {
            sshagent (sshCreds) {
                withCredentials(creds) {
                    body()
                }
            }
        } else {
            withCredentials(creds) {
                body()
            }
        }
    } else {
        body()
    }
}

// reused in sonarExecuteScan
void handleErrorDetails(String stepName, Closure body) {
    try {
        body()
    } catch (ex) {
        def errorDetailsFileName = "${stepName}_errorDetails.json"
        if (fileExists(file: errorDetailsFileName)) {
            def errorDetails = readJSON(file: errorDetailsFileName)
            def errorCategory = ""
            if (errorDetails.category) {
                errorCategory = " (category: ${errorDetails.category})"
                DebugReport.instance.failedBuild.category = errorDetails.category
            }
            error "[${stepName}] Step execution failed${errorCategory}. Error: ${errorDetails.error?:errorDetails.message}"
        }
        error "[${stepName}] Step execution failed. Error: ${ex}, please see log file for more details."
    }
}
