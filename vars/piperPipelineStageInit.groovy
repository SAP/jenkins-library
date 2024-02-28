import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.JenkinsUtils
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
     * Defines the library resource containing the legacy configuration definition.
     */
    'legacyConfigSettings',
    /**
     * Defines the main branch for your pipeline. **Typically this is the `master` branch, which does not need to be set explicitly.** Only change this in exceptional cases to a fixed branch name.
     */
    'productiveBranch',
    /**
     * Name of the project, e.g. used for the name of lockable resources.
     */
    'projectName',
    /**
     * Specify to execute artifact versioning in a kubernetes pod.
     * @possibleValues `true`, `false`
     */
    'runArtifactVersioningOnPod',
    /**
     *  Defines the library resource containing stage/step initialization settings. Those define conditions when certain steps/stages will be activated. **Caution: changing the default will break the standard behavior of the pipeline - thus only relevant when including `Init` stage into custom pipelines!**
     */
    'stageConfigResource',
    /**
     * Defines the library resource containing the stash settings to be performed before and after each stage. **Caution: changing the default will break the standard behavior of the pipeline - thus only relevant when including `Init` stage into custom pipelines!**
     */
    'stashSettings',
    /**
    * Works as the stashSettings parameter, but allows the use of a stash settings file that is not available as a library resource.
    */
    'customStashSettings',
    /**
     * Whether verbose output should be produced.
     * @possibleValues `true`, `false`
     */
    'verbose'
]
@Field STAGE_STEP_KEYS = [
    /**
     * Sets the build version.
     * @possibleValues `true`, `false`
     */
    'artifactPrepareVersion',
    /**
     * Retrieve transport request from git commit history.
     * @possibleValues `true`, `false`
     */
    'transportRequestReqIDFromGit'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /**
     * Enables the use of technical stage names.
     */
    'useTechnicalStageNames',
    /**
     * Provides a clone from the specified repository.
     * This map contains attributes, such as, `branches`, `extensions`, `userRemoteConfigs` etc.
     * Example: `[$class: 'GitSCM', branches: [[name: <branch_to_be_cloned>]], userRemoteConfigs: [[credentialsId: <credential_to_access_repository>, url: <repository_url>]]]`.
     */
    'checkoutMap',
    /**
     * The map returned from a Jenkins git checkout. Used to set the git information in the
     * common pipeline environment.
     */
    'scmInfo',
    /**
     * Optional skip of checkout if checkout was done before this step already.
     * @possibleValues `true`, `false`
     */
    'skipCheckout',
    /**
    * Mandatory if you skip the checkout. Then you need to unstash your workspace to get the e.g. configuration.
    */
    'stashContent',
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
        def skipCheckout = parameters.skipCheckout
        if (skipCheckout != null && !(skipCheckout instanceof Boolean)) {
            error "[${STEP_NAME}] Parameter skipCheckout has to be of type boolean. Instead got '${skipCheckout.class.getName()}'"
        }
        def scmInfo = parameters.scmInfo
        if (skipCheckout && !scmInfo) {
            error "[${STEP_NAME}] Need am scmInfo map retrieved from a checkout. " +
                "If you want to skip the checkout the scm info needs to be provided by you with parameter scmInfo, " +
                "for example as follows:\n" +
                "  def scmInfo = checkout scm\n" +
                "  piperPipelineStageInit script:this, skipCheckout: true, scmInfo: scmInfo"
        }
        if (!skipCheckout) {
            scmInfo = checkout(parameters.checkoutMap ?: scm)
        }
        else {
            def stashContent = parameters.stashContent
            if(stashContent == null || stashContent.size() == 0) {
                error "[${STEP_NAME}] needs stashes if you skip checkout"
            }
            utils.unstashAll(stashContent)
        }

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
            checkForLegacyConfiguration(script: script, legacyConfigSettings: legacyConfigSettings)
        }

        String buildTool = config.buildTool
        String buildToolDesc = inferBuildToolDesc(script, config.buildTool)

        checkBuildTool(buildTool, buildToolDesc)

        script.commonPipelineEnvironment.projectName = config.projectName

        if (!script.commonPipelineEnvironment.projectName && config.inferProjectName) {
            script.commonPipelineEnvironment.projectName = inferProjectName(script, buildTool, buildToolDesc)
        }

        if (Boolean.valueOf(env.ON_K8S) && config.containerMapResource) {
            ContainerMap.instance.initFromResource(script, config.containerMapResource, buildTool)
        }

        initStashConfiguration(script, config.stashSettings, config.customStashSettings, config.verbose ?: false)

        if (config.verbose) {
            echo "piper-lib-os  configuration: ${script.commonPipelineEnvironment.configuration}"
        }

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

            config.artifactPrepareVersion = true
        }

        if (config.artifactPrepareVersion) {
            Map prepareVersionParams = [script: script]
            if (config.inferBuildTool) {
                prepareVersionParams.buildTool = buildTool
            }
            if(buildToolDesc) {
                prepareVersionParams.filePath = buildToolDesc
            }
            if (env.ON_K8S && !config.runArtifactVersioningOnPod) {
                // We force dockerImage: "" for the K8S case to avoid the execution of artifactPrepareVersion in a K8S Pod.
                // Since artifactPrepareVersion may need the ".git" folder in order to push a tag, it would need to be part of the stashing.
                // There are however problems with tar-ing this folder, which results in a failure to copy the stash back -- without a failure of the pipeline.
                // This then also has the effect that any changes made to the build descriptors by the step (updated version) are not visible in the relevant stashes.
                // In addition, a mvn executable is available on the Jenkins instance which can be used directly instead of executing the command in a container.
                prepareVersionParams.dockerImage = ""
            }
            artifactPrepareVersion prepareVersionParams
        }

        if (config.transportRequestReqIDFromGit) {
            transportRequestReqIDFromGit(script: script)
        }
        pipelineStashFilesBeforeBuild script: script
    }
}

// Infer build tool descriptor (maven, npm, mta)
private static String inferBuildToolDesc(script, buildTool) {

    String buildToolDesc = null

    switch (buildTool) {
        case 'maven':
            Map configBuild = script.commonPipelineEnvironment.getStepConfiguration('mavenBuild', 'Build')
            buildToolDesc = configBuild.pomPath? configBuild.pomPath : 'pom.xml'
            break
        case 'npm': // no parameter for the descriptor path
            buildToolDesc = 'package.json'
            break
        case 'mta':
            Map configBuild = script.commonPipelineEnvironment.getStepConfiguration('mtaBuild', 'Build')
            buildToolDesc = configBuild.source? configBuild.source + '/mta.yaml' : 'mta.yaml'
            break
        default:
            break;
    }

    return buildToolDesc
}

private String inferProjectName(Script script, String buildTool, String buildToolDesc) {
    switch (buildTool) {
        case 'maven':
            def pom = script.readMavenPom file: buildToolDesc
            return "${pom.groupId}-${pom.artifactId}"
        case 'npm':
            Map packageJson = script.readJSON file: buildToolDesc
            return packageJson.name
        case 'mta':
            Map mta = script.readYaml file: buildToolDesc
            return mta.ID
    }

    script.error "Cannot infer projectName. Project buildTool was none of the expected ones 'mta', 'maven', or 'npm'."
}

private checkBuildTool(String buildTool, String buildDescriptorPattern) {
    if (buildTool != "mta" && !findFiles(glob: buildDescriptorPattern)) {
        error "[${STEP_NAME}] buildTool configuration '${buildTool}' does not fit to your project (buildDescriptorPattern: '${buildDescriptorPattern}'), please set buildTool as general setting in your .pipeline/config.yml correctly, see also https://sap.github.io/jenkins-library/configuration/"
    }
}

private void initStashConfiguration (script, stashSettings, customStashSettings, verbose) {
    Map stashConfiguration = null
    if (customStashSettings){
        stashConfiguration = readYaml(file: customStashSettings)
    }else{
        stashConfiguration = readYaml(text: libraryResource(stashSettings))
    }
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
