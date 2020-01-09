import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils

import groovy.transform.Field

import static com.sap.piper.Utils.downloadSettingsFromUrl

@Field def STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = [
    /** @see dockerExecute */
    'dockerImage',
    /** Path or url to the mvn settings file that should be used as global settings file.*/
    'globalSettingsFile',
    /** Path or url to the mvn settings file that should be used as project settings file.*/
    'projectSettingsFile',
    /** Path to the pom file that should be used.*/
    'pomPath',
    /** Path to the location of the local repository that should be used.*/
    'm2Path'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /** @see dockerExecute */
    'dockerOptions',
    /** Flags to provide when running mvn.*/
    'flags',
    /** Maven goals that should be executed.*/
    'goals',
    /** Additional properties.*/
    'defines',
    /**
     * Configures maven to log successful downloads. This is set to `false` by default to reduce the noise in build logs.
     * @possibleValues `true`, `false`
     */
    'logSuccessfulMavenTransfers'
])

/**
 * Executes a maven command inside a Docker container.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        final script = checkScript(this, parameters) ?: this

        // load default & individual configuration
        Map configuration = ConfigurationHelper.newInstance(this, script)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], configuration)

        String command = "mvn"

        String globalSettingsFile = configuration.globalSettingsFile?.trim()
        if (globalSettingsFile) {
            if (globalSettingsFile.startsWith("http")) {
                globalSettingsFile = downloadSettingsFromUrl(this, globalSettingsFile, 'global-settings.xml')
            }
            command += " --global-settings '${globalSettingsFile}'"
        }

        String m2Path = configuration.m2Path
        if(m2Path?.trim()) {
            command += " -Dmaven.repo.local='${m2Path}'"
        }

        String projectSettingsFile = configuration.projectSettingsFile?.trim()
        if (projectSettingsFile) {
            if (projectSettingsFile.startsWith("http")) {
                projectSettingsFile = downloadSettingsFromUrl(this, projectSettingsFile, 'project-settings.xml')
            }
            command += " --settings '${projectSettingsFile}'"
        }

        String pomPath = configuration.pomPath
        if(pomPath?.trim()){
            command += " --file '${pomPath}'"
        }

        String mavenFlags = configuration.flags
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
