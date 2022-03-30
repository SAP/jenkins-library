import com.sap.piper.BashUtils
import com.sap.piper.DebugReport
import com.sap.piper.DefaultValueCache
import com.sap.piper.JenkinsUtils
import com.sap.piper.MapUtils
import com.sap.piper.PiperGoUtils
import com.sap.piper.Utils
import com.sap.piper.analytics.InfluxData
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

void call(Map parameters = [:], String stepName, String metadataFile, List credentialInfo, boolean failOnMissingReports = false, boolean failOnMissingLinks = false, boolean failOnError = false) {

    Map handlePipelineStepErrorsParameters = [stepName: stepName, stepParameters: parameters]
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
        echo "Step params $stepParameters"

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

            //Add ANS credential information to the config
            config += ["ansServiceKeyCredentialsId": script.commonPipelineEnvironment.configuration.hooks.ansServiceKeyCredentialsId]

            // prepare stashes
            // first eliminate empty stashes
            config.stashContent = utils.unstashAll(config.stashContent)
            // then make sure that commonPipelineEnvironment, config, ... is also available when step stashing is active
            if (config.stashContent?.size() > 0) {
                config.stashContent.add('pipelineConfigAndTests')
                config.stashContent.add('piper-bin')
                config.stashContent.add('pipelineStepReports')
            }

            if (parameters.stashNoDefaultExcludes) {
                // Merge this parameter which is only relevant in Jenkins context
                // (for dockerExecuteOnKubernetes step) and go binary doesn't know about
                config.stashNoDefaultExcludes = parameters.stashNoDefaultExcludes
            }

            dockerWrapper(script, stepName, config) {
                handleErrorDetails(stepName) {
                    writePipelineEnv(script: script, piperGoPath: piperGoPath)
                    utils.unstash('pipelineStepReports')
                    try {
                        try {
                            try {
                                credentialWrapper(config, credentialInfo) {
                                    sh "${piperGoPath} ${stepName}${defaultConfigArgs}${customConfigArg}"
                                }
                            } finally {
                                jenkinsUtils.handleStepResults(stepName, failOnMissingReports, failOnMissingLinks)
                            }
                        } finally {
                           readPipelineEnv(script: script, piperGoPath: piperGoPath)
                        }
                    } finally {
                        InfluxData.readFromDisk(script)
                        stash name: 'pipelineStepReports', includes: '.pipeline/stepReports/**', allowEmpty: true
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
    credentialInfo = handleVaultCredentials(config, credentialInfo)
    credentialInfo = handleANSCredentials(config, credentialInfo)
    if (credentialInfo.size() > 0) {
        def creds = []
        def sshCreds = []
        credentialInfo.each { cred ->
            def credentialsId
            if (cred.resolveCredentialsId == false) {
                credentialsId = cred.id
            } else {
                credentialsId = config[cred.id]
            }
            if (credentialsId) {
                switch (cred.type) {
                    case "file":
                        creds.add(file(credentialsId: credentialsId, variable: cred.env[0]))
                        break
                    case "token":
                        creds.add(string(credentialsId: credentialsId, variable: cred.env[0]))
                        break
                    case "usernamePassword":
                        creds.add(usernamePassword(credentialsId: credentialsId, usernameVariable: cred.env[0], passwordVariable: cred.env[1]))
                        break
                    case "ssh":
                        sshCreds.add(credentialsId)
                        break
                    default:
                        error("invalid credential type: ${cred.type}")
                }
            }
        }

        // remove credentialIds that were probably defaulted and which are not present in jenkins
        if (containsVaultConfig(config)) {
            creds = removeMissingCredentials(creds, config)
            sshCreds = removeMissingCredentials(sshCreds, config)
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

List removeMissingCredentials(List creds, Map config) {
    return creds.findAll { credentialExists(it, config) }
}

boolean credentialExists(cred, Map config) {
    try {
        withCredentials([cred]) {
            return true
        }
    } catch (e) {
        return false
    }
}

boolean containsVaultConfig(Map config) {
    def approleIsUsed = config.containsKey('vaultAppRoleTokenCredentialsId') && config.containsKey('vaultAppRoleSecretTokenCredentialsId')
    def tokenIsUsed = config.containsKey('vaultTokenCredentialsId')

    return approleIsUsed || tokenIsUsed
}

// Injects vaultCredentials if steps supports resolving parameters from vault
List handleVaultCredentials(config, List credentialInfo) {
    if (config.containsKey('vaultAppRoleTokenCredentialsId') && config.containsKey('vaultAppRoleSecretTokenCredentialsId')) {
        credentialInfo += [[type: 'token', id: 'vaultAppRoleTokenCredentialsId', env: ['PIPER_vaultAppRoleID']],
                            [type: 'token', id: 'vaultAppRoleSecretTokenCredentialsId', env: ['PIPER_vaultAppRoleSecretID']]]
    }

    if (config.containsKey('vaultTokenCredentialsId')) {
        credentialInfo += [[type: 'token', id: 'vaultTokenCredentialsId', env: ['PIPER_vaultToken']]]
    }

    return credentialInfo
}

List handleANSCredentials(config, List credentialInfo){
    if (config.containsKey('ansServiceKeyCredentialsId')) {
        credentialInfo += [[type: 'token', id: 'ansServiceKeyCredentialsId', env: ['PIPER_ansServiceKey']]]
    }

    return credentialInfo
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
            error "[${stepName}] Step execution failed${errorCategory}. Error: ${errorDetails.error ?: errorDetails.message}"
        }
        error "[${stepName}] Step execution failed. Error: ${ex}, please see log file for more details."
    }
}
