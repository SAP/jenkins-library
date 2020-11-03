import com.cloudbees.groovy.cps.NonCPS

import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GitUtils
import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import com.sap.piper.analytics.InfluxData

import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /** */
    'collectTelemetryData',

    /** Credentials (username and password) used to download custom defaults if access is secured.*/
    'customDefaultsCredentialsId',

    /** Enable automatic inference of build tool (maven, npm, mta) based on existing project files.
     * If this is set to true, it is not required to set the build tool by hand for those cases.
     */
    'inferBuildTool'
]

@Field Set STEP_CONFIG_KEYS = []

@Field Set PARAMETER_KEYS = [
    /** Path to the pipeline configuration file defining project specific settings.*/
    'configFile',
    /** A list of file names which will be extracted from library resources and which serve as source for
     * default values for the pipeline configuration. These are merged with and override built-in defaults, with
     * a parameter supplied by the last resource file taking precedence over the same parameter supplied in an
     * earlier resource file or built-in default.*/
    'customDefaults',
    /** A list of file paths or URLs which must point to YAML content. These work exactly like
     * `customDefaults`, but from local or remote files instead of library resources. They are merged with and
     * take precedence over `customDefaults`.*/
    'customDefaultsFromFiles',
    /**
      * The projects git repo url. Typically the fetch url.
      */
    'gitUrl',
]

/**
 * Initializes the [`commonPipelineEnvironment`](commonPipelineEnvironment.md), which is used throughout the complete pipeline.
 *
 * !!! tip
 *     This step needs to run at the beginning of a pipeline right after the SCM checkout.
 *     Then subsequent pipeline steps consume the information from `commonPipelineEnvironment`; it does not need to be passed to pipeline steps explicitly.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters)

        def gitUtils = parameters.gitUtils ?: new GitUtils()

        String configFile = parameters.get('configFile')
        loadConfigurationFromFile(script, configFile)

        // Copy custom defaults from library resources to include them in the 'pipelineConfigAndTests' stash
        List customDefaultsResources = Utils.appendParameterToStringList(
            ['default_pipeline_environment.yml'], parameters, 'customDefaults')
        customDefaultsResources.each {
            cd ->
                writeFile file: ".pipeline/${cd}", text: libraryResource(cd)
        }

        List customDefaultsFiles = Utils.appendParameterToStringList(
            [], parameters, 'customDefaultsFromFiles')

        if (script.commonPipelineEnvironment.configuration.customDefaults) {
            if (!script.commonPipelineEnvironment.configuration.customDefaults in List) {
                // Align with Go side on supported parameter type.
                error "You have defined the parameter 'customDefaults' in your project configuration " +
                    "but it is of an unexpected type. Please make sure that it is a list of strings, i.e. " +
                    "customDefaults = ['...']. See https://sap.github.io/jenkins-library/configuration/ for " +
                    "more details."
            }
            customDefaultsFiles = Utils.appendParameterToStringList(
                customDefaultsFiles, script.commonPipelineEnvironment.configuration as Map, 'customDefaults')
        }
        String customDefaultsCredentialsId = script.commonPipelineEnvironment.configuration.general?.customDefaultsCredentialsId
        customDefaultsFiles = copyOrDownloadCustomDefaultsIntoPipelineEnv(script, customDefaultsFiles, customDefaultsCredentialsId)

        prepareDefaultValues([
            script: script,
            customDefaults: parameters.customDefaults,
            customDefaultsFromFiles: customDefaultsFiles ])

        piperLoadGlobalExtensions script: script, customDefaults: parameters.customDefaults, customDefaultsFromFiles: customDefaultsFiles

        stash name: 'pipelineConfigAndTests', includes: '.pipeline/**', allowEmpty: true

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        inferBuildTool(script, config)

        (parameters.utils ?: new Utils()).pushToSWA([
            step: STEP_NAME,
            stepParamKey4: 'customDefaults',
            stepParam4: parameters.customDefaults?'true':'false'
        ], config)

        InfluxData.addField('step_data', 'build_url', env.BUILD_URL)
        InfluxData.addField('pipeline_data', 'build_url', env.BUILD_URL)

        def gitCommitId = gitUtils.getGitCommitIdOrNull()
        if (gitCommitId) {
            script.commonPipelineEnvironment.setGitCommitId(gitCommitId)
        }

        if (config.gitUrl) {
            setGitUrlsOnCommonPipelineEnvironment(script, config.gitUrl)
        }
    }
}


// Infer build tool (maven, npm, mta) based on existing build descriptor files in the project root.
private static void inferBuildTool(script, config) {
    // For backwards compatibility, build tool inference must be enabled via inferBuildTool setting
    boolean inferBuildTool = config?.inferBuildTool

    if (inferBuildTool) {
        boolean isMtaProject = script.fileExists('mta.yaml')
        def isMavenProject = script.fileExists('pom.xml')
        def isNpmProject = script.fileExists('package.json')

        if (isMtaProject) {
            script.commonPipelineEnvironment.buildTool = 'mta'
        } else if (isMavenProject) {
            script.commonPipelineEnvironment.buildTool = 'maven'
        } else if (isNpmProject) {
            script.commonPipelineEnvironment.buildTool = 'npm'
        }
    }
}

private static loadConfigurationFromFile(script, String configFile) {
    if (!configFile) {
        String defaultYmlConfigFile = '.pipeline/config.yml'
        String defaultYamlConfigFile = '.pipeline/config.yaml'
        if (script.fileExists(defaultYmlConfigFile)) {
            configFile = defaultYmlConfigFile
        } else if (script.fileExists(defaultYamlConfigFile)) {
            configFile = defaultYamlConfigFile
        }
    }

    // A file passed to the function is not checked for existence in order to fail the pipeline.
    if (configFile) {
        script.commonPipelineEnvironment.configuration = script.readYaml(file: configFile)
        script.commonPipelineEnvironment.configurationFile = configFile
    }
}

private static List copyOrDownloadCustomDefaultsIntoPipelineEnv(script, List customDefaults, String credentialsId) {
    List fileList = []
    int urlCount = 0
    for (int i = 0; i < customDefaults.size(); i++) {
        // copy retrieved file to .pipeline/ to make sure they are in the pipelineConfigAndTests stash
        if (!(customDefaults[i] in CharSequence) || customDefaults[i] == '') {
            script.echo "WARNING: Ignoring invalid entry in custom defaults from files: '${customDefaults[i]}'"
            continue
        }
        String fileName
        if (customDefaults[i].startsWith('http://') || customDefaults[i].startsWith('https://')) {
            fileName = "custom_default_from_url_${urlCount}.yml"

            Map httpRequestParameters = [
                url: customDefaults[i],
                validResponseCodes: '100:399,404' // Allow a more specific error message for 404 case
            ]
            if (credentialsId) {
                httpRequestParameters.authentication = credentialsId
            }
            def response = script.httpRequest(httpRequestParameters)
            if (response.status == 404) {
                error "URL for remote custom defaults (${customDefaults[i]}) appears to be incorrect. " +
                    "Server returned HTTP status code 404. " +
                    "Please make sure that the path is correct and no authentication is required to retrieve the file."
            }

            script.writeFile file: ".pipeline/$fileName", text: response.content
            urlCount++
        } else if (script.fileExists(customDefaults[i])) {
            fileName = customDefaults[i]
            script.writeFile file: ".pipeline/$fileName", text: script.readFile(file: fileName)
        } else {
            script.echo "WARNING: Custom default entry not found: '${customDefaults[i]}', it will be ignored"
            continue
        }
        fileList.add(fileName)
    }
    return fileList
}

/*
 * Returns the parts of an url.
 * Valid keys for the retured map are:
 *   - protocol
 *   - auth
 *   - host
 *   - port
 *   - path
 */
@NonCPS
/* private */ Map parseUrl(String url) {

    def urlMatcher = url =~ /^((http|https|git|ssh):\/\/)?((.*)@)?([^:\/]+)(:([\d]*))?(\/?(.*))$/

    return [
        protocol: urlMatcher[0][2],
        auth: urlMatcher[0][4],
        host: urlMatcher[0][5],
        port: urlMatcher[0][7],
        path: urlMatcher[0][9],
    ]
}

private void setGitUrlsOnCommonPipelineEnvironment(script, String gitUrl) {

    Map url = parseUrl(gitUrl)

    if (url.protocol in ['http', 'https']) {
        script.commonPipelineEnvironment.setGitSshUrl("git@${url.host}:${url.path}")
        script.commonPipelineEnvironment.setGitHttpsUrl(gitUrl)
    } else if (url.protocol in [ null, 'ssh', 'git']) {
        script.commonPipelineEnvironment.setGitSshUrl(gitUrl)
        script.commonPipelineEnvironment.setGitHttpsUrl("https://${url.host}/${url.path}")
    }

    List gitPathParts = url.path.replaceAll('.git', '').split('/')
    def gitFolder = 'N/A'
    def gitRepo = 'N/A'
    switch (gitPathParts.size()) {
        case 0:
            break
        case 1:
            gitRepo = gitPathParts[0]
            break
        case 2:
            gitFolder = gitPathParts[0]
            gitRepo = gitPathParts[1]
            break
        default:
            gitRepo = gitPathParts[gitPathParts.size()-1]
            gitPathParts.remove(gitPathParts.size()-1)
            gitFolder = gitPathParts.join('/')
            break
    }
    script.commonPipelineEnvironment.setGithubOrg(gitFolder)
    script.commonPipelineEnvironment.setGithubRepo(gitRepo)
}
