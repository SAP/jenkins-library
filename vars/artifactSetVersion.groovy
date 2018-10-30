import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.GitUtils
import com.sap.piper.Utils
import com.sap.piper.versioning.ArtifactVersioning

import groovy.transform.Field
import groovy.text.SimpleTemplateEngine

@Field String STEP_NAME = 'artifactSetVersion'
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

        def script = checkScript(this, parameters)

        def gitUtils = parameters.juStabGitUtils ?: new GitUtils()

        if (gitUtils.insideWorkTree()) {
            if (sh(returnStatus: true, script: 'git diff --quiet HEAD') != 0)
                error "[${STEP_NAME}] Files in the workspace have been changed previously - aborting ${STEP_NAME}"
        }
        if (script == null)
            script = this
        // load default & individual configuration
        ConfigurationHelper configHelper = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, this, CONFIG_KEY_COMPATIBILITY)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, this, CONFIG_KEY_COMPATIBILITY)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS, this, CONFIG_KEY_COMPATIBILITY)
            .mixin(gitCommitId: gitUtils.getGitCommitIdOrNull())
            .mixin(parameters, PARAMETER_KEYS, this, CONFIG_KEY_COMPATIBILITY)
            .withMandatoryProperty('buildTool')
            .dependingOn('buildTool').mixin('filePath')
            .dependingOn('buildTool').mixin('versioningTemplate')

        Map config = configHelper.use()

        config = configHelper.addIfEmpty('timestamp', getTimestamp(config.timestampTemplate))
                             .use()

        new Utils().pushToSWA([step: STEP_NAME, stepParam1: config.buildTool, stepParam2: config.artifactType, stepParam3: parameters?.script == null], config)

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
            config = new ConfigurationHelper(config)
                .addIfEmpty('gitSshUrl', isAppContainer(config)
                            ?script.commonPipelineEnvironment.getAppContainerProperty('gitSshUrl')
                            :script.commonPipelineEnvironment.getGitSshUrl())
                .withMandatoryProperty('gitSshUrl')
                .use()
            
            def gitConfig = []

            if(config.gitUserEMail) gitConfig.add("-c user.email=\"${config.gitUserEMail}\"")
            if(config.gitUserName)  gitConfig.add("-c user.name=\"${config.gitUserName}\"")
            gitConfig = gitConfig.join(' ')

            try {
                sh """#!/bin/bash
                      git add .
                      git ${gitConfig} commit -m 'update version ${newVersion}'
                      git tag ${config.tagPrefix}${newVersion}"""
                config.gitCommitId = gitUtils.getGitCommitIdOrNull()
            } catch (e) {
                error "[${STEP_NAME}]git commit and tag failed: ${e}"
            }

            sshagent([config.gitSshKeyCredentialsId]) {
                sh "git push ${config.gitSshUrl} ${config.tagPrefix}${newVersion}"
            }
        }

        if (isAppContainer(config)) {
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

def isAppContainer(config){
    return config.buildTool == 'docker' && config.artifactType == 'appContainer'
}

def getTimestamp(pattern){
    return sh(returnStdout: true, script: "date --universal +'${pattern}'").trim()
}
