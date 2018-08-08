import com.sap.piper.ConfigurationHelper
import com.sap.piper.GitUtils
import com.sap.piper.Utils
import com.sap.piper.versioning.ArtifactVersioning

import groovy.transform.Field
import groovy.text.SimpleTemplateEngine

@Field String STEP_NAME = 'artifactSetVersion'
@Field Set GENERAL_CONFIG_KEYS = ['collectTelemetryData']
@Field Map CONFIG_KEY_COMPATIBILITY = [gitSshKeyCredentialsId: 'gitCredentialsId']
@Field Set STEP_CONFIG_KEYS = [
    'artifactType',
    'buildTool',
    'commitVersion',
    'dockerVersionSource',
    'filePath',
    'gitSshKeyCredentialsId',
    'gitUserEMail',
    'gitUserName',
    'gitSshUrl',
    'tagPrefix',
    'timestamp',
    'timestampTemplate',
    'versioningTemplate'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus('gitCommitId')

def call(Map parameters = [:], Closure body = null) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

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
            .mixinGeneralConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, this, CONFIG_KEY_COMPATIBILITY)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, this, CONFIG_KEY_COMPATIBILITY)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS, this, CONFIG_KEY_COMPATIBILITY)
            .mixin(gitCommitId: gitUtils.getGitCommitIdOrNull())
            .mixin(parameters, PARAMETER_KEYS, this, CONFIG_KEY_COMPATIBILITY)
            .withMandatoryProperty('buildTool')
            .dependingOn('buildTool').mixin('filePath')
            .dependingOn('buildTool').mixin('versioningTemplate')
            .use()

        config = new ConfigurationHelper(config)
            .addIfEmpty('gitSshUrl', (config.buildTool == 'docker' && config.artifactType == 'appContainer')?script.commonPipelineEnvironment.getAppContainerProperty('gitSshUrl'):script.commonPipelineEnvironment.getGitSshUrl())
            .addIfEmpty('timestamp', getTimestamp(config.timestampTemplate))
            .withMandatoryProperty('gitSshUrl')
            .use()

        new Utils().pushToSWA([step: STEP_NAME, stepParam1: config.buildTool], config)

        def artifactVersioning = ArtifactVersioning.getArtifactVersioning(config.buildTool, script, config)
        def currentVersion = artifactVersioning.getVersion()

        def newVersion
        if (config.artifactType == 'appContainer' && config.dockerVersionSource == 'appVersion'){
            newVersion = currentVersion
        } else {
            def binding = [version: currentVersion, timestamp: config.timestamp, commitId: config.gitCommitId]
            newVersion = new SimpleTemplateEngine().createTemplate(config.versioningTemplate).make(binding).toString()
        }

        artifactVersioning.setVersion(newVersion)

        if(body != null){
            body(newVersion)
        }

        if (config.commitVersion) {
            sh 'git add .'

            sshagent([config.gitSshKeyCredentialsId]) {
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

                config.gitCommitId = gitUtils.getGitCommitIdOrNull()
            }
        }

        if (config.buildTool == 'docker' && config.artifactType == 'appContainer') {
            script.commonPipelineEnvironment.setAppContainerProperty('artifactVersion', newVersion)
            script.commonPipelineEnvironment.setAppContainerProperty('gitCommitId', config.gitCommitId)
        } else {
            //standard case
            script.commonPipelineEnvironment.setArtifactVersion(newVersion)
            script.commonPipelineEnvironment.setGitCommitId(config.gitCommitId)
        }

        echo "[${STEP_NAME}]New version: ${newVersion}"
    }
}

def getTimestamp(pattern){
    return sh(returnStdout: true, script: "date --universal +'${pattern}'").trim()
}
