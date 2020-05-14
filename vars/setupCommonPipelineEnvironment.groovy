import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import com.sap.piper.analytics.InfluxData

import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    /** */
    'collectTelemetryData'
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
    'customDefaultsFromFiles'
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
        customDefaultsFiles = copyOrDownloadCustomDefaultsIntoPipelineEnv(script, customDefaultsFiles)

        prepareDefaultValues([
            script: script,
            customDefaults: parameters.customDefaults,
            customDefaultsFromFiles: customDefaultsFiles ])

        stash name: 'pipelineConfigAndTests', includes: '.pipeline/**', allowEmpty: true

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .use()

        (parameters.utils ?: new Utils()).pushToSWA([
            step: STEP_NAME,
            stepParamKey4: 'customDefaults',
            stepParam4: parameters.customDefaults?'true':'false'
        ], config)

        InfluxData.addField('step_data', 'build_url', env.BUILD_URL)
        InfluxData.addField('pipeline_data', 'build_url', env.BUILD_URL)
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

private static List copyOrDownloadCustomDefaultsIntoPipelineEnv(script, List customDefaults) {
    List fileList = []
    int urlCount = 0
    for (int i = 0; i < customDefaults.size(); i++) {
        // copy retrieved file to .pipeline/ to make sure they are in the pipelineConfigAndTests stash
        String fileName
        if (customDefaults[i].startsWith('http://') || customDefaults[i].startsWith('https://')) {
            fileName = "custom_default_from_url_${urlCount}.yml"

            def response = script.httpRequest(
                url: customDefaults[i],
                validResponseCodes: '100:399,404' // Allow a more specific error message for 404 case
            )
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
