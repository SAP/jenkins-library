import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper

import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /** Allows to overwrite the GitHub API url.*/
    'githubApiUrl',
    /**
     * Allows to overwrite the GitHub token credentials id.
     * @possibleValues Jenkins credential id
     */
    'githubTokenCredentialsId',
    /** Allows to overwrite the GitHub url.*/
    'githubServerUrl'
]

@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * If it is set to `true`, a list of all closed issues and merged pull-requests since the last release will added below the `releaseBodyHeader`.
     * @possibleValues `true`, `false`
     */
    'addClosedIssues',
    /**
     * If you set `addDeltaToLastRelease` to `true`, a link will be added to the relese information that brings up all commits since the last release.
     * @possibleValues `true`, `false`
     */
    'addDeltaToLastRelease',
    /** Allows to pass additional filter criteria for retrieving closed issues since the last release. Additional criteria could be for example specific `label`, or `filter` according to [GitHub API documentation](https://developer.github.com/v3/issues/).*/
    'customFilterExtension',
    /** Allows to exclude issues with dedicated labels. Usage is like `excludeLabels: ['label1', 'label2']`.*/
    'excludeLabels',
    /** Allows to overwrite the GitHub organitation.*/
    'githubOrg',
    /** Allows to overwrite the GitHub repository.*/
    'githubRepo',
    /** Allows to specify the content which will appear for the release.*/
    'releaseBodyHeader',
    /** Defines the version number which will be written as tag as well as release name.*/
    'version'
])

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * This step creates a tag in your GitHub repository together with a release.
 *
 * The release can be filled with text plus additional information like:
 *
 * * Closed pull request since last release
 * * Closed issues since last release
 * * link to delta information showing all commits since last release
 *
 * The result looks like
 *
 * ![Example release](../images/githubRelease.png)
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .addIfEmpty('githubOrg', script.commonPipelineEnvironment.getGithubOrg())
            .addIfEmpty('githubRepo', script.commonPipelineEnvironment.getGithubRepo())
            .addIfEmpty('version', script.commonPipelineEnvironment.getArtifactVersion())
            .withMandatoryProperty('githubOrg')
            .withMandatoryProperty('githubRepo')
            .withMandatoryProperty('githubTokenCredentialsId')
            .withMandatoryProperty('version')
            .use()

        new Utils().pushToSWA([step: STEP_NAME], config)

        withCredentials([string(credentialsId: config.githubTokenCredentialsId, variable: 'TOKEN')]) {
            def releaseBody = config.releaseBodyHeader?"${config.releaseBodyHeader}<br />":''
            def content = getLastRelease(config, TOKEN)
            if (config.addClosedIssues)
                releaseBody += addClosedIssue(config, TOKEN, content.published_at)
            if (config.addDeltaToLastRelease)
                releaseBody += addDeltaToLastRelease(config, content.tag_name)
            postNewRelease(config, TOKEN, releaseBody)
        }
    }
}

Map getLastRelease(config, TOKEN){
    def result = [:]

    def response = httpRequest "${config.githubApiUrl}/repos/${config.githubOrg}/${config.githubRepo}/releases/latest?access_token=${TOKEN}"
    if (response.status == 200) {
        result = readJSON text: response.content
    } else {
        echo "[${STEP_NAME}] This is the first release - no previous releases available"
        config.addDeltaToLastRelease = false
    }
    return result
}

String addClosedIssue(config, TOKEN, publishedAt){
    if (config.customFilterExtension) {
        config.customFilterExtension = "&${config.customFilterExtension}"
    }

    def publishedAtFilter = publishedAt ? "&since=${publishedAt}": ''

    def response = httpRequest "${config.githubApiUrl}/repos/${config.githubOrg}/${config.githubRepo}/issues?access_token=${TOKEN}&per_page=100&state=closed&direction=asc${publishedAtFilter}${config.customFilterExtension}"
    def result = ''

    content = readJSON text: response.content

    //list closed pull-requests
    result += '<br />**List of closed pull-requests since last release**<br />'
    for (def item : content) {
        if (item.pull_request && !isExcluded(item, config.excludeLabels)) {
            result += "[#${item.number}](${item.html_url}): ${item.title}<br />"
        }
    }
    //list closed issues
    result += '<br />**List of closed issues since last release**<br />'
    for (def item : content) {
        if (!item.pull_request && !isExcluded(item, config.excludeLabels)) {
            result += "[#${item.number}](${item.html_url}): ${item.title}<br />"
        }
    }
    return result
}

String addDeltaToLastRelease(config, latestTag){
    def result = ''
    //add delta link to previous release
    result += '<br />**Changes**<br />'
    result += "[${latestTag}...${config.version}](${config.githubServerUrl}/${config.githubOrg}/${config.githubRepo}/compare/${latestTag}...${config.version}) <br />"
    return result
}

void postNewRelease(config, TOKEN, releaseBody){
    releaseBody = releaseBody.replace('"', '\\"')
    //write release information
    def data = "{\"tag_name\": \"${config.version}\",\"target_commitish\": \"master\",\"name\": \"${config.version}\",\"body\": \"${releaseBody}\",\"draft\": false,\"prerelease\": false}"
    try {
        httpRequest httpMode: 'POST', requestBody: data, url: "${config.githubApiUrl}/repos/${config.githubOrg}/${config.githubRepo}/releases?access_token=${TOKEN}"
    } catch (e) {
        echo """[${STEP_NAME}] Error occured when writing release information
---------------------------------------------------------------------
Request body was:
---------------------------------------------------------------------
${data}
---------------------------------------------------------------------"""
        throw e
    }
}

boolean isExcluded(item, excludeLabels){
    def result = false
    excludeLabels.each {labelName ->
        item.labels.each { label ->
            if (label.name == labelName) {
                result = true
            }
        }
    }
    return result
}
