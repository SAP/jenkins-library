import com.sap.piper.ConfigurationHelper
import com.sap.piper.GitUtils
import com.sap.piper.Utils
import com.sap.piper.versioning.ArtifactVersioning

import groovy.transform.Field
import groovy.text.SimpleTemplateEngine

@Field String STEP_NAME = 'artifactSetVersion'
@Field Set GENERAL_CONFIG_KEYS = ['collectTelemetryData']
@Field Set STEP_CONFIG_KEYS = [
    'artifactType',
    'buildTool',
    'commitVersion',
    'dockerVersionSource',
    'filePath',
    'gitCredentialsId',
    'gitUserEMail',
    'gitUserName',
    'gitSshUrl',
    'tagPrefix',
    'timestamp',
    'timestampTemplate',
    'versioningTemplate'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus('gitCommitId')

def call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        // def utils = parameters.juStabUtils ?: new Utils()
        def gitUtils = parameters.juStabGitUtils ?: new GitUtils()

        if (fileExists('.git')) {
            if (sh(returnStatus: true, script: 'git diff --quiet HEAD') != 0)
                error "[${STEP_NAME}] Files in the workspace have been changed previously - aborting ${STEP_NAME}"
        }

        def script = parameters.script
        if (script == null)
            script = this

        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(gitCommitId: gitUtils.getGitCommitIdOrNull())
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('buildTool')
            .use()

        new Utils().pushToSWA([step: STEP_NAME, stepParam1: config.buildTool], config)

        if (!config.filePath)
            config.filePath = config[config.buildTool].filePath //use default configuration

        def newVersion
        def artifactVersioning = ArtifactVersioning.getArtifactVersioning(config.buildTool, script, config)

        if(config.artifactType == 'appContainer' && config.dockerVersionSource == 'appVersion'){
            if (script.commonPipelineEnvironment.getArtifactVersion())
                //replace + sign if available since + is not allowed in a Docker tag
                newVersion = script.commonPipelineEnvironment.getArtifactVersion().replace('+', '_')
            else
                error ("[${STEP_NAME}] No artifact version available for 'dockerVersionSource: appVersion' -> executeBuild needs to run for the application artifact first to set the artifactVersion for the application artifact.'")
        } else {
            def currentVersion = artifactVersioning.getVersion()

            def timestamp = config.timestamp ? config.timestamp : getTimestamp(config.timestampTemplate)

            def versioningTemplate = config.versioningTemplate ? config.versioningTemplate : config[config.buildTool].versioningTemplate
            //defined in default configuration
            def binding = [version: currentVersion, timestamp: timestamp, commitId: config.gitCommitId]
            def templatingEngine = new SimpleTemplateEngine()
            def template = templatingEngine.createTemplate(versioningTemplate).make(binding)
            newVersion = template.toString()
        }

        artifactVersioning.setVersion(newVersion)

        def gitCommitId

        if (config.commitVersion) {
            sh 'git add .'

            sshagent([config.gitCredentialsId]) {
                def gitUserMailConfig = ''
                if (config.gitUserName && config.gitUserEMail)
                    gitUserMailConfig = "-c user.email=\"${config.gitUserEMail}\" -c user.name=\"${config.gitUserName}\""

                try {
                    sh "git ${gitUserMailConfig} commit -m 'update version ${newVersion}'"
                } catch (e) {
                    error "[${STEP_NAME}]git commit failed: ${e}"
                }
                sh "git remote set-url origin ${config.gitSshUrl}"
                sh "git tag ${config.tagPrefix}${newVersion}"
                sh "git push origin ${config.tagPrefix}${newVersion}"

                gitCommitId = gitUtils.getGitCommitIdOrNull()
            }
        }

        if (config.buildTool == 'docker' && config.artifactType == 'appContainer') {
            script.commonPipelineEnvironment.setAppContainerProperty('artifactVersion', newVersion)
            script.commonPipelineEnvironment.setAppContainerProperty('gitCommitId', gitCommitId)
        } else {
            //standard case
            script.commonPipelineEnvironment.setArtifactVersion(newVersion)
            script.commonPipelineEnvironment.setGitCommitId(gitCommitId)
        }

        echo "[${STEP_NAME}]New version: ${newVersion}"
    }
}

def getTimestamp(pattern){
    return sh(returnStdout: true, script: "date --universal +'${pattern}'").trim()
}
