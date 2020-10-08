import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.JenkinsUtils
import com.sap.piper.LegacyConfigurationCheckUtils
import com.sap.piper.StageNameProvider
import com.sap.piper.Utils
import com.sap.piper.k8s.ContainerMap
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String TECHNICAL_STAGE_NAME = 'init'

@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Defines the build tool used.
     * @possibleValues `docker`, `kaniko`, `maven`, `mta, ``npm`
     */
    'buildTool',
    /**
     * Defines the library resource containing the container map.
     */
    'containerMapResource',
    /**
     * Enable automatic inference of build tool (maven, npm, mta) based on existing project files.
     * If this is set to true, it is not required to provide the `buildTool` parameter in the `general` section of the pipeline configuration.
     */
    'inferBuildTool',
    /**
     * Enables automatic inference from the build descriptor in case projectName is not configured.
     */
    'inferProjectName',
    /**
     * Toggle for initialization of the stash settings for Cloud SDK Pipeline.
     * If this is set to true, the stashSettings parameter is **not** configurable.
     */
    'initCloudSdkStashSettings',
    /**
     * Defines the library resource containing the legacy configuration definition.
     */
    'legacyConfigSettings',
    /**
     * Defines the main branch for your pipeline. **Typically this is the `master` branch, which does not need to be set explicitly.** Only change this in exceptional cases
     */
    'productiveBranch',
    /**
     * Name of the project, e.g. used for the name of lockable resources.
     */
    'projectName',
    /**
     * Defines the library resource containing stage/step initialization settings. Those define conditions when certain steps/stages will be activated. **Caution: changing the default will break the standard behavior of the pipeline - thus only relevant when including `Init` stage into custom pipelines!**
     */
    'stageConfigResource',
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
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /**
     * Enables the use of technical stage names.
     */
    'useTechnicalStageNames',
    /**
     * Optional path to the pipeline configuration file defining project specific settings.
     */
    'configFile',
    /**
     * Optional list of file names which will be extracted from library resources and which serve as source for
     * default values for the pipeline configuration. These are merged with and override built-in defaults, with
     * a parameter supplied by the last resource file taking precedence over the same parameter supplied in an
     * earlier resource file or built-in default.
     */
    'customDefaults',
    /**
     * Optional list of file paths or URLs which must point to YAML content. These work exactly like
     * `customDefaults`, but from local or remote files instead of library resources. They are merged with and
     * take precedence over `customDefaults`.
     */
    'customDefaultsFromFiles'
])

/**
 * This stage initializes the pipeline run and prepares further execution.
 *
 * It will check out your repository and perform some steps to initialize your pipeline run.
 */
@GenerateStageDocumentation(defaultStageName = 'Init')
void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    if (parameters.useTechnicalStageNames) {
        StageNameProvider.instance.useTechnicalStageNames = true
    }

    def stageName = StageNameProvider.instance.getStageName(script, parameters, this)

    piperStageWrapper (script: script, stageName: stageName, stashContent: [], ordinal: 1, telemetryDisabled: true) {
        def scmInfo = checkout scm

        setupCommonPipelineEnvironment(script: script, customDefaults: parameters.customDefaults, scmInfo: scmInfo,
            configFile: parameters.configFile, customDefaultsFromFiles: parameters.customDefaultsFromFiles)

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .addIfEmpty('stageConfigResource', 'com.sap.piper/pipeline/stageDefaults.yml')
            .addIfEmpty('stashSettings', 'com.sap.piper/pipeline/stashSettings.yml')
            .addIfEmpty('buildTool', script.commonPipelineEnvironment.buildTool)
            .withMandatoryProperty('buildTool')
            .use()

        if (config.legacyConfigSettings) {
            Map legacyConfigSettings = readYaml(text: libraryResource(config.legacyConfigSettings))
            LegacyConfigurationCheckUtils.checkConfiguration(script, legacyConfigSettings)
        }

        String buildTool = checkBuildTool(config)

        script.commonPipelineEnvironment.projectName = config.projectName

        if (!script.commonPipelineEnvironment.projectName && config.inferProjectName) {
            script.commonPipelineEnvironment.projectName = inferProjectName(script, buildTool)
        }

        if (Boolean.valueOf(env.ON_K8S) && config.containerMapResource) {
            ContainerMap.instance.initFromResource(script, config.containerMapResource, buildTool)
        }

        //perform stashing based on library resource piper-stash-settings.yml if not configured otherwise or Cloud SDK Pipeline is initialized
        if (config.initCloudSdkStashSettings) {
            switch (buildTool) {
                case 'maven':
                    initStashConfiguration(script, "com.sap.piper/pipeline/cloudSdkJavaStashSettings.yml", config.verbose?: false)
                    break
                case 'npm':
                    initStashConfiguration(script, "com.sap.piper/pipeline/cloudSdkJavascriptStashSettings.yml", config.verbose?: false)
                    break
                case 'mta':
                    initStashConfiguration(script, "com.sap.piper/pipeline/cloudSdkMtaStashSettings.yml", config.verbose?: false)
                    break
                default:
                    error "[${STEP_NAME}] No stash settings for build tool ${buildTool} can be found. With `initCloudSdkStashSettings` active, only Maven, MTA or NPM projects are supported."
                    break
            }
        } else {
            initStashConfiguration(script, config.stashSettings, config.verbose?: false)
        }

        if (config.verbose) {
            echo "piper-lib-os  configuration: ${script.commonPipelineEnvironment.configuration}"
        }

        // telemetry reporting
        utils.pushToSWA([step: STEP_NAME], config)

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
            if (config.inferBuildTool && env.ON_K8S) {
                // We set dockerImage: "" for the K8S case to avoid the execution of artifactPrepareVersion in a K8S Pod.
                // In addition, a mvn executable is available on the Jenkins instance which can be used directly instead of executing the command in a container.
                artifactPrepareVersion script: script, buildTool: buildTool, dockerImage: ""
            } else if (config.inferBuildTool) {
                artifactPrepareVersion script: script, buildTool: buildTool
            } else {
                artifactSetVersion script: script
            }
        }
        pipelineStashFilesBeforeBuild script: script
    }
}

private String inferProjectName(Script script, String buildTool) {
    switch (buildTool) {
        case 'maven':
            def pom = script.readMavenPom file: 'pom.xml'
            return "${pom.groupId}-${pom.artifactId}"
        case 'npm':
            Map packageJson = script.readJSON file: 'package.json'
            return packageJson.name
        case 'mta':
            Map mta = script.readYaml file: 'mta.yaml'
            return mta.ID
    }

    script.error "Cannot infer projectName. Project buildTool was none of the expected ones 'mta', 'maven', or 'npm'."
}

private String checkBuildTool(config) {
    def buildDescriptorPattern = ''
    String buildTool = config.buildTool

    switch (buildTool) {
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
        error "[${STEP_NAME}] buildTool configuration '${config.buildTool}' does not fit to your project, please set buildTool as genereal setting in your .pipeline/config.yml correctly, see also https://sap.github.io/jenkins-library/configuration/"
    }
    return buildTool
}

private void initStashConfiguration (script, stashSettings, verbose) {
    Map stashConfiguration = readYaml(text: libraryResource(stashSettings))
    if (verbose) echo "Stash config: ${stashConfiguration}"
    script.commonPipelineEnvironment.configuration.stageStashes = stashConfiguration
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
