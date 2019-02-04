import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    'buildTool',
    'productiveBranch',
    'stashSettings',
    'verbose'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    def stageName = parameters.stageName?:env.STAGE_NAME

    piperStageWrapper (script: script, stageName: stageName, stashContent: [], ordinal: 1) {
        def scmInfo = checkout scm

        setupCommonPipelineEnvironment script: script, customDefaults: parameters.customDefaults

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .addIfEmpty('stageConfigResource', 'com.sap.piper/pipeline/stageDefaults.yml')
            .addIfEmpty('stashSettings', 'com.sap.piper/pipeline/stashSettings.yml')
            .withMandatoryProperty('buildTool')
            .use()

        //perform stashing based on libray resource piper-stash-settings.yml if not configured otherwise
        initStashConfiguration(script, config)

        setScmInfoOnCommonPipelineEnvironment(script, scmInfo)
        script.commonPipelineEnvironment.setGitCommitId(scmInfo.GIT_COMMIT)

        if (config.verbose) {
            echo "piper-lib-os  configuration: ${script.commonPipelineEnvironment.configuration}"
        }

        // telemetry reporting
        utils.pushToSWA([step: STEP_NAME], config)

        checkBuildTool(config)

        piperInitRunStageConfiguration script: script, stageConfigResource: config.stageConfigResource

        if (env.BRANCH_NAME == config.productiveBranch) {
            artifactSetVersion script: script
        }

        pipelineStashFilesBeforeBuild script: script

    }
}

private void checkBuildTool(config) {
    def buildDescriptorPattern = ''
    switch (config.buildTool) {
        case 'maven':
            buildDescriptorPattern = 'pom.xml'
            break
        case 'npm':
            buildDescriptorPattern = 'package.json'
            break
        case 'mta':
            buildDescriptorPattern = 'mta.yaml'
            break
    }
    if (buildDescriptorPattern && !findFiles(glob: buildDescriptorPattern)) {
        error "[${STEP_NAME}] buildTool configuration '${config.buildTool}' does not fit to your project, please set buildTool as genereal setting in your .pipeline/config.yml correctly, see also https://github.wdf.sap.corp/pages/ContinuousDelivery/piper-doc/configuration/"
    }
}

private void initStashConfiguration (script, config) {
    Map stashConfiguration = readYaml(text: libraryResource(config.stashSettings))
    echo "Stash config: stashConfiguration"
    script.commonPipelineEnvironment.configuration.stageStashes = stashConfiguration
}

private void setScmInfoOnCommonPipelineEnvironment(script, scmInfo) {

    def gitUrl = scmInfo.GIT_URL

    if (gitUrl.startsWith('http')) {
        def httpPattern = /(https?):\/\/([^:\/]+)(?:[:\d\/]*)(.*)/
        def gitMatcher = gitUrl =~ httpPattern
        if (!gitMatcher.hasGroup() && gitMatcher.groupCount() != 3) return
        script.commonPipelineEnvironment.setGitSshUrl("git@${gitMatcher[0][2]}:${gitMatcher[0][3]}")
        script.commonPipelineEnvironment.setGitHttpsUrl(gitUrl)
    } else if (gitUrl.startsWith('ssh')) {
        //(.*)@([^:\/]*)(?:[:\d\/]*)(.*)
        def httpPattern = /(.*)@([^:\/]*)(?:[:\d\/]*)(.*)/
        def gitMatcher = gitUrl =~ httpPattern
        if (!gitMatcher.hasGroup() && gitMatcher.groupCount() != 3) return
        script.commonPipelineEnvironment.setGitSshUrl(gitUrl)
        script.commonPipelineEnvironment.setGitHttpsUrl("https://${gitMatcher[0][2]}/${gitMatcher[0][3]}")
    }
    else if (gitUrl.indexOf('@') > 0) {
        script.commonPipelineEnvironment.setGitSshUrl(gitUrl)
        script.commonPipelineEnvironment.setGitHttpsUrl("https://${(gitUrl.split('@')[1]).replace(':', '/')}")
    }
}
