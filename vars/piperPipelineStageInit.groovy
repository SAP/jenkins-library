import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Defines the build tool used.
     * @possibleValues `docker`, `kaniko`, `maven`, `mta, ``npm`
     */
    'buildTool',
    /**
     * Defines the main branch for your pipeline. **Typically this is the `master` branch, which does not need to be set explicitly.** Only change this in exceptional cases
     */
    'productiveBranch',
    /**
     * Defines the library resource containing the stash settings to be performed before and after each stage. **Caution: changing the default will break the standard behavior of the pipeline - thus only relevant when including `Init` stage into custom pipelines!**
     */
    'stashSettings',
    /**
     * Whether verbose output should be produced.
     * @possibleValues `true`, `false`
     */
    'verbose'
]
@Field STAGE_STEP_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * This stage initializes the pipeline run and prepares further execution.
 *
 * It will check out your repository and perform some steps to initialize your pipeline run.
 */
@GenerateStageDocumentation(defaultStageName = 'Init')
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
            .addIfEmpty('stageOrdinals', 'com.sap.piper/pipeline/stageOrdinals.yml')
            //.addIfEmpty('stashSettings', 'com.sap.piper/pipeline/stashSettings.yml')
            .withMandatoryProperty('buildTool')
            .use()

        //perform stashing based on libray resource piper-stash-settings.yml if not configured otherwise
        initStashConfiguration(script, config)

        //provide the correct ordinals for stage locking in piperStageWrapper
       // initStageOrdinalConfiguration (script, config)

        setScmInfoOnCommonPipelineEnvironment(script, scmInfo)
        script.commonPipelineEnvironment.setGitCommitId(scmInfo.GIT_COMMIT)

        if (config.verbose) {
            echo "piper-lib-os  configuration: ${script.commonPipelineEnvironment.configuration}"
        }

        // telemetry reporting
        utils.pushToSWA([step: STEP_NAME], config)

        checkBuildTool(config)

        piperInitRunStageConfiguration script: script, stageConfigResource: config.stageConfigResource

        // CHANGE_ID is set only for pull requests
        if (env.CHANGE_ID) {
            List prActions = []

            //get trigger action from comment like /piper action
            def jenkinsUtils = new JenkinsUtils()
            def commentTriggerAction = jenkinsUtils.getIssueCommentTriggerAction()

            if (commentTriggerAction != null) prActions.add(commentTriggerAction)

            try {
                prActions.addAll(pullRequest.getLabels().asList())
            } catch (ex) {
                echo "[${STEP_NAME}] GitHub labels could not be retrieved from Pull Request, please make sure that credentials are maintained on multi-branch job."
            }


            setPullRequestStageStepActivation(script, config, prActions)
        }

        if (env.BRANCH_NAME == config.productiveBranch) {
            if (parameters.script.commonPipelineEnvironment.configuration.runStep?.get('Init')?.slackSendNotification) {
                slackSendNotification script: script, message: "STARTED: Job <${env.BUILD_URL}|${URLDecoder.decode(env.JOB_NAME, java.nio.charset.StandardCharsets.UTF_8.name())} ${env.BUILD_DISPLAY_NAME}>", color: 'WARNING'
            }
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
    if (config.verbose) echo "Stash config: ${stashConfiguration}"
    script.commonPipelineEnvironment.configuration.stageStashes = stashConfiguration
}

private void initStageOrdinalConfiguration (script, config) {
    Map stageOrdinals = readYaml(text: libraryResource(config.stageOrdinals))
    if (config.verbose) echo "Stage ordinals: ${stageOrdinals}"
    if (script.commonPipelineEnvironment.configuration.general == null) script.commonPipelineEnvironment.configuration.general = [:]
    script.commonPipelineEnvironment.configuration.general += stageOrdinals
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

private void setPullRequestStageStepActivation(script, config, List actions) {

    if (script.commonPipelineEnvironment.configuration.runStep == null)
        script.commonPipelineEnvironment.configuration.runStep = [:]
    if (script.commonPipelineEnvironment.configuration.runStep[config.pullRequestStageName] == null)
        script.commonPipelineEnvironment.configuration.runStep[config.pullRequestStageName] = [:]

    actions.each {action ->
        if (action.startsWith(config.labelPrefix))
            action = action.minus(config.labelPrefix)

        def stepName = config.stepMappings[action]
        if (stepName) {
            script.commonPipelineEnvironment.configuration.runStep."${config.pullRequestStageName}"."${stepName}" = true
        }
    }
}
