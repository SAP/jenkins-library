import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.GitUtils
import com.sap.piper.Utils
import com.sap.piper.versioning.ArtifactVersioning

import groovy.transform.Field
import groovy.text.SimpleTemplateEngine

@Field String STEP_NAME = getClass().getName()
@Field Map CONFIG_KEY_COMPATIBILITY = [gitSshKeyCredentialsId: 'gitCredentialsId']

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

@Field Set STEP_CONFIG_KEYS = [
    /** Defines the type of the artifact.
    * @possibleValues `appContainer`
    */
    'artifactType',
    /** Defines the tool which is used for building the artifact.
    * @possibleValues docker, dlang, golang, maven, mta, npm, pip, sbt
    */
    'buildTool',
    /** Controls if the changed version is committed and pushed to the git repository.
    * If this is enabled (which is the default), you need to provide `gitCredentialsId` and `gitSshUrl`.
    * @possibleValues `true`, `false`
    */
    'commitVersion',
    /** Specifies the source to be used for the main version which is used for generating the automatic version.
    *  * This can either be the version of the base image - as retrieved from the `FROM` statement within the Dockerfile, e.g. `FROM jenkins:2.46.2`
    *  * Alternatively the name of an environment variable defined in the Docker image can be used which contains the version number, e.g. `ENV MY_VERSION 1.2.3`
    *  * The third option `appVersion` applies only to the artifactType `appContainer`. Here the version of the app which is packaged into the container will be used as version for the container itself.
    * @possibleValues FROM, (ENV name),appVersion
    */
    'dockerVersionSource',
    /** Defines a custom path to the descriptor file.*/
    'filePath',
    /** Defines the ssh git credentials to be used for writing the tag.*/
    'gitSshKeyCredentialsId',
    /** Allows to overwrite the global git setting 'user.email' available on your Jenkins server.*/
    'gitUserEMail',
    /** Allows to overwrite the global git setting 'user.name' available on your Jenkins server.*/
    'gitUserName',
    /** Defines the git ssh url to the source code repository.*/
    'gitSshUrl',
    /** Defines the prefix which is used for the git tag which is written during the versioning run.*/
    'tagPrefix',
    /** Defines the timestamp to be used in the automatic version string. You could overwrite the default behavior by explicitly setting this string.*/
    'timestamp',
    /** */
    'timestampTemplate',
    /** */
    'versioningTemplate'
]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus(
    /** Defines the version prefix of the automatically generated version. By default it will take the long commitId hash.
    * You could pass any other string (e.g. the short commitId hash) to be used. In case you don't want to have the gitCommitId added to the automatic versioning string you could set the value to an empty string: `''`.
    */
    'gitCommitId'
    )

/** The continuous delivery process requires that each build is done with a unique version number.
* 
* The version generated using this step will contain:
* 
* * Version (major.minor.patch) from descriptor file in master repository is preserved. Developers should be able to autonomously decide on increasing either part of this version number.
* * Timestamp
* * CommitId (by default the long version of the hash)
* 
* Optionally, but enabled by default, the new version is pushed as a new tag into the source code repository (e.g. GitHub).
* If this option is chosen, git credentials and the repository URL needs to be provided.
* Since you might not want to configure the git credentials in Jenkins, committing and pushing can be disabled using the `commitVersion` parameter as described below.
* If you require strict reproducibility of your builds, this should be used.
*/
@GenerateDocumentation
void call(Map parameters = [:], Closure body = null) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters)

        def gitUtils = parameters.juStabGitUtils ?: new GitUtils()

        if (gitUtils.isWorkTreeDirty()) {
                error "[${STEP_NAME}] Files in the workspace have been changed previously - aborting ${STEP_NAME}"
        }
        if (script == null)
            script = this
        // load default & individual configuration
        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixin(gitCommitId: gitUtils.getGitCommitIdOrNull())
            .mixin(parameters, PARAMETER_KEYS, CONFIG_KEY_COMPATIBILITY)
            .withMandatoryProperty('buildTool')
            .dependingOn('buildTool').mixin('filePath')
            .dependingOn('buildTool').mixin('versioningTemplate')

        Map config = configHelper.use()

        config = configHelper.addIfEmpty('timestamp', getTimestamp(config.timestampTemplate))
                             .use()

        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'buildTool',
            stepParam1: config.buildTool,
            stepParamKey2: 'artifactType',
            stepParam2: config.artifactType,
            stepParamKey3: 'scriptMissing',
            stepParam3: parameters?.script == null
        ], config)

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
            config = ConfigurationHelper.newInstance(this, config)
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
