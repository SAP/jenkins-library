import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils

import groovy.transform.Field

@Field def STEP_NAME = 'mavenExecute'

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = [
    'dockerImage',
    'globalSettingsFile',
    'projectSettingsFile',
    'pomPath',
    'm2Path'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    'dockerOptions',
    'flags',
    'goals',
    'defines',
    'logSuccessfulMavenTransfers'
])

def call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        final script = parameters.script

        // load default & individual configuration
        Map configuration = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        new Utils().pushToSWA([step: STEP_NAME], configuration)

        String command = "mvn"

        def globalSettingsFile = configuration.globalSettingsFile
        if (globalSettingsFile?.trim()) {
            if(globalSettingsFile.trim().startsWith("http")){
                downloadSettingsFromUrl(globalSettingsFile)
                globalSettingsFile = "settings.xml"
            }
            command += " --global-settings '${globalSettingsFile}'"
        }

        def m2Path = configuration.m2Path
        if(m2Path?.trim()) {
            command += " -Dmaven.repo.local='${m2Path}'"
        }

        def projectSettingsFile = configuration.projectSettingsFile
        if (projectSettingsFile?.trim()) {
            if(projectSettingsFile.trim().startsWith("http")){
                downloadSettingsFromUrl(projectSettingsFile)
                projectSettingsFile = "settings.xml"
            }
            command += " --settings '${projectSettingsFile}'"
        }

        def pomPath = configuration.pomPath
        if(pomPath?.trim()){
            command += " --file '${pomPath}'"
        }

        def mavenFlags = configuration.flags
        if (mavenFlags?.trim()) {
            command += " ${mavenFlags}"
        }

        // Always use Maven's batch mode
        if (!(command =~ /--batch-mode|-B(?=\s)|-B\\|-B$/)) {
            command += ' --batch-mode'
        }

        // Disable log for successful transfers by default. Note this requires the batch-mode flag.
        final String disableSuccessfulMavenTransfersLogFlag = ' -Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn'
        if (!configuration.logSuccessfulMavenTransfers) {
            if (!command.contains(disableSuccessfulMavenTransfersLogFlag)) {
                command += disableSuccessfulMavenTransfersLogFlag
            }
        }

        def mavenGoals = configuration.goals
        if (mavenGoals?.trim()) {
            command += " ${mavenGoals}"
        }
        def defines = configuration.defines
        if (defines?.trim()){
            command += " ${defines}"
        }
        dockerExecute(script: script, dockerImage: configuration.dockerImage, dockerOptions: configuration.dockerOptions) {
            sh command
        }
    }
}

private downloadSettingsFromUrl(String url){
    def settings = httpRequest url
    writeFile file: 'settings.xml', text: settings.getContent()
}

