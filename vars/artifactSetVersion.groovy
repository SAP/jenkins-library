import static com.sap.piper.Prerequisites.checkScript
import static com.sap.piper.BashUtils.quoteAndEscape as q

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.GitUtils
import com.sap.piper.Utils
import com.sap.piper.versioning.ArtifactVersioning

import groovy.transform.Field
import groovy.text.GStringTemplateEngine

enum GitPushMode {NONE, HTTPS, SSH}

@Field String STEP_NAME = getClass().getName()
@Field Map CONFIG_KEY_COMPATIBILITY = [gitSshKeyCredentialsId: 'gitCredentialsId']

@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS

@Field Set STEP_CONFIG_KEYS = [
    /**
     * Defines the type of the artifact.
     * @possibleValues `appContainer`
     */
    'artifactType',
    /**
     * Defines the tool which is used for building the artifact.
     * @possibleValues `dub`, `docker`, `golang`, `maven`, `mta`, `npm`, `pip`, `sbt`
     */
    'buildTool',
    /**
     * Controls if the changed version is committed and pushed to the git repository.
     * If this is enabled (which is the default), you need to provide `gitCredentialsId` and `gitSshUrl`.
     * @possibleValues `true`, `false`
     */
    'commitVersion',
    /**
      * Prints some more information for troubleshooting. May reveal security relevant information. Usage is recommanded for troubleshooting only. Productive usage
      * is not recommended.
      * @possibleValues `true`, `false`
      */
    'verbose',
    /**
     * Specifies the source to be used for the main version which is used for generating the automatic version.
     * * This can either be the version of the base image - as retrieved from the `FROM` statement within the Dockerfile, e.g. `FROM jenkins:2.46.2`
     * * Alternatively the name of an environment variable defined in the Docker image can be used which contains the version number, e.g. `ENV MY_VERSION 1.2.3`
     * * The third option `appVersion` applies only to the artifactType `appContainer`. Here the version of the app which is packaged into the container will be used as version for the container itself.
     * @possibleValues FROM, (ENV name),appVersion
     */
    'dockerVersionSource',
    /**
     * Defines a custom path to the descriptor file.
     */
    'filePath',
    /**
     * Defines the ssh git credentials to be used for writing the tag.
     */
    'gitSshKeyCredentialsId',
    /** */
    'gitHttpsCredentialsId',
    /**
     * Allows to overwrite the global git setting 'user.email' available on your Jenkins server.
     */
    'gitUserEMail',
    /**
     * Allows to overwrite the global git setting 'user.name' available on your Jenkins server.
     */
    'gitUserName',
    /**
     * Defines the git ssh url to the source code repository. Used in conjunction with 'GitPushMode.SSH'.
     * @mandatory for `gitPushMode` `SSH`
     */
    'gitSshUrl',
    /**
     * Defines the git https url to the source code repository. Used in conjunction with 'GitPushMode.HTTPS'.
     * @mandatory for `gitPushMode` `HTTPS`
     */
    'gitHttpsUrl',
    /**
     * Disables the ssl verification for git push. Intended to be used only for troubleshooting. Productive usage is not recommanded.
     */
    'gitDisableSslVerification',
    /**
     * Defines the prefix which is used for the git tag which is written during the versioning run.
     */
    'tagPrefix',
    /**
     * Defines the timestamp to be used in the automatic version string. You could overwrite the default behavior by explicitly setting this string.
     */
    'timestamp',
    /** Defines the template for the timestamp which will be part of the created version. */
    'timestampTemplate',
    /** Defines the template for the automatic version which will be created. */
    'versioningTemplate',
    /** Controls which protocol is used for performing push operation to remote repo.
      * Required credentials needs to be configured ('gitSshKeyCredentialsId'/'gitHttpsCredentialsId').
      * Push is only performed in case 'commitVersion' is set to 'true'.
      * @possibleValues 'SSH', 'HTTPS', 'NONE'
      */
    'gitPushMode'
]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus(
    /**
     * Defines the version prefix of the automatically generated version. By default it will take the long commitId hash.
     * You could pass any other string (e.g. the short commitId hash) to be used. In case you don't want to have the gitCommitId added to the automatic versioning string you could set the value to an empty string: `''`.
     */
    'gitCommitId'
)

/**
 * The continuous delivery process requires that each build is done with a unique version number.
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

        def script = checkScript(this, parameters) ?: this
        String stageName = parameters.stageName ?: env.STAGE_NAME

        def gitUtils = parameters.juStabGitUtils ?: new GitUtils()
        if (gitUtils.isWorkTreeDirty()) {
            error "[${STEP_NAME}] Files in the workspace have been changed previously - aborting ${STEP_NAME}"
        }

        // load default & individual configuration
        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixin(gitCommitId: gitUtils.getGitCommitIdOrNull())
            .mixin(parameters, PARAMETER_KEYS, CONFIG_KEY_COMPATIBILITY)
            .withMandatoryProperty('buildTool')
            .dependingOn('buildTool').mixin('filePath')
            .dependingOn('buildTool').mixin('versioningTemplate')

        Map config = configHelper.use()

        GitPushMode gitPushMode = config.gitPushMode

        config = configHelper.addIfEmpty('timestamp', getTimestamp(config.timestampTemplate))
            .use()

        def artifactVersioning = ArtifactVersioning.getArtifactVersioning(config.buildTool, script, config)
        def currentVersion = artifactVersioning.getVersion()

        def newVersion
        if (config.artifactType == 'appContainer' && config.dockerVersionSource == 'appVersion'){
            newVersion = currentVersion
        } else {
            def binding = [version: currentVersion, timestamp: config.timestamp, commitId: config.gitCommitId]
            newVersion = new GStringTemplateEngine().createTemplate(config.versioningTemplate).make(binding).toString()
        }

        artifactVersioning.setVersion(newVersion)

        if(body != null){
            body(newVersion)
        }

        if (config.commitVersion) {

            def gitConfig = []

            if(config.gitUserEMail) {
                gitConfig.add("-c user.email=${q(config.gitUserEMail)}")
            } else {
                // in case there is no user.email configured on project level we might still
                // be able to work in case there is a configuration available on plain git level.
                if(sh(returnStatus: true, script: 'git config user.email') != 0) {
                    error 'No git user.email configured. Neither via project config nor on plain git level.'
                }
            }
            if(config.gitUserName) {
                gitConfig.add("-c user.name=${q(config.gitUserName)}")
            } else {
                // in case there is no user.name configured on project level we might still
                // be able to work in case there is a configuration available on plain git level.
                if (sh(returnStatus: true, script: 'git config user.name') != 0) {
                    error 'No git user.name configured. Neither via project config nor on plain git level.'
                }
            }
            gitConfig = gitConfig.join(' ')

            try {
                sh """#!/bin/bash
                    set -e
                    git add . --update
                    git ${gitConfig} commit -m 'update version ${newVersion}'
                    git tag ${q(config.tagPrefix+newVersion)}"""
                config.gitCommitId = gitUtils.getGitCommitIdOrNull()
            } catch (e) {
                error "[${STEP_NAME}]git commit and tag failed: ${e}"
            }

            if(gitPushMode == GitPushMode.SSH) {

                config = ConfigurationHelper.newInstance(this, config)
                    .addIfEmpty('gitSshUrl', isAppContainer(config)
                                ?script.commonPipelineEnvironment.getAppContainerProperty('gitSshUrl')
                                :script.commonPipelineEnvironment.getGitSshUrl())
                    .withMandatoryProperty('gitSshUrl')
                    .use()

                sshagent([config.gitSshKeyCredentialsId]) {
                    sh "git push ${q(config.gitSshUrl)} ${q(config.tagPrefix+newVersion)}"
                }

            } else if(gitPushMode == GitPushMode.HTTPS) {

                config = ConfigurationHelper.newInstance(this, config)
                    .addIfEmpty('gitSshUrl', isAppContainer(config)
                                ?script.commonPipelineEnvironment.getAppContainerProperty('gitHttpsUrl')
                                :script.commonPipelineEnvironment.getGitHttpsUrl())
                    .withMandatoryProperty('gitHttpsUrl')
                    .use()

                withCredentials([usernamePassword(
                    credentialsId: config.gitHttpsCredentialsId,
                    passwordVariable: 'PASSWORD',
                    usernameVariable: 'USERNAME')]) {

                    // Problem: when username/password is encoded and in case the encoded version differs from
                    // the non-encoded version  (e.g. '@'  gets replaced by '%40') the encoded version
                    // it is not replaced by stars in the log by surrounding withCredentials.
                    // In order to avoid having the secrets in the log we take the following actions in case
                    // the encoded version(s) differs from the non-encoded versions
                    //
                    // 1.) we switch off '-x' in the hashbang
                    // 2.) we tell git push to be silent
                    // 3.) we send stderr to /dev/null
                    //
                    // Disadvantage: In this case we don't see any output for troubleshooting.

                    def USERNAME_ENCODED = URLEncoder.encode(USERNAME, 'UTF-8'),
                        PASSWORD_ENCODED = URLEncoder.encode(PASSWORD, 'UTF-8')

                    boolean encodedVersionsDiffers = USERNAME_ENCODED != USERNAME || PASSWORD_ENCODED != PASSWORD

                    def prefix = 'https://'
                    def gitUrlWithCredentials = config.gitHttpsUrl.replaceAll("^${prefix}", "${prefix}${USERNAME_ENCODED}:${PASSWORD_ENCODED}@")

                    def hashbangFlags = '-xe'
                    def gitPushFlags = []
                    def streamhandling = ''
                    def gitDebug = ''
                    gitConfig = []

                    if(config.gitHttpProxy) {
                        gitConfig.add("-c http.proxy=${q(config.gitHttpProxy)}")
                    }

                    if(config.gitDisableSslVerification) {
                        echo 'git ssl verification is switched off. This setting is not recommanded in productive environments.'
                        gitConfig.add('-c http.sslVerify=false')
                    }

                    if(encodedVersionsDiffers) {
                        if(config.verbose) { // known issue: in case somebody provides the stringish 'false' we get the boolean value 'true' here.
                            echo 'Verbose flag set, but encoded username/password differs from unencoded version. Cannot provide verbose output in this case. ' +
                                    'In order to enable verbose output switch to a username/password which is not altered by url encoding.'
                        }
                        hashbangFlags = '-e'
                        streamhandling ='&>/dev/null'
                        gitPushFlags.add('--quiet')
                        echo 'Performing git push in quiet mode.'
                    } else {
                        if(config.verbose) { // known issue: in case somebody provides the stringish 'false' we get the boolean value 'true' here.
                            echo 'Verbose mode enabled. This is not recommanded for productive usage. This might reveal security sensitive information.'
                            gitDebug ='git config --list; env |grep proxy; GIT_CURL_VERBOSE=1 GIT_TRACE=1 '
                            gitPushFlags.add('--verbose')
                        }
                    }

                    gitConfig = gitConfig.join(' ')
                    gitPushFlags = gitPushFlags.join(' ')

                    sh script:   """|#!/bin/bash ${hashbangFlags}
                                    |${gitDebug}git ${gitConfig} push ${gitPushFlags} ${gitUrlWithCredentials} ${q(config.tagPrefix+newVersion)} ${streamhandling}""".stripMargin()
                }
            } else {
                echo "Git push mode: ${gitPushMode.toString()}. Git push to remote has been skipped."
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
    return sh(returnStdout: true, script: "date --utc +${q(pattern)}").trim()
}
