import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.variablesubstitution.ExecutionContext
import com.sap.piper.variablesubstitution.DebugHelper
import com.sap.piper.variablesubstitution.YamlUtils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS + [
    /**
     * The `String` path of the Yaml file to replace variables in.
     * Defaults to "manifest.yml" if not specified otherwise.
     */
    'manifestFile',
    /**
     * The `String` path of the Yaml file to produce as output.
     * If not specified this will default to `manifestFile` and overwrite it.
     */
    'outputManifestFile',
    /**
     * The `List` of `String` paths of the Yaml files containing the variable values to use as a replacement in the manifest file.
     * Defaults to `["manifest-variables.yml"]` if not specified otherwise. The order of the files given in the list is relevant
     * in case there are conflicting variable names and values within variable files. In such a case, the values of the last file win.
     */
    'manifestVariablesFiles',
    /**
     * A `List` of `Map` entries for key-value pairs used for variable substitution within the file given by `manifestFile`.
     * Defaults to an empty list, if not specified otherwise. This can be used to set variables like it is provided
     * by `cf push --var key=value`.
     *
     * The order of the maps of variables given in the list is relevant in case there are conflicting variable names and values
     * between maps contained within the list. In case of conflicts, the last specified map in the list will win.
     *
     * Though each map entry in the list can contain more than one key-value pair for variable substitution, it is recommended
     * to stick to one entry per map, and rather declare more maps within the list. The reason is that
     * if a map in the list contains more than one key-value entry, and the entries are conflicting, the
     * conflict resolution behavior is undefined (since map entries have no sequence).
     *
     * Note: variables defined via `manifestVariables` always win over conflicting variables defined via any file given
     * by `manifestVariablesFiles` - no matter what is declared before. This reproduces the same behavior as can be
     * observed when using `cf push --var` in combination with `cf push --vars-file`.
     */
    'manifestVariables'
]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Step to substitute variables in a given YAML file with those specified in one or more variables files given by the
 * `manifestVariablesFiles` parameter. This follows the behavior of `cf push --vars-file`, and can be
 * used as a pre-deployment step if commands other than `cf push` are used for deployment (e.g. `cf blue-green-deploy`).
 *
 * The format to reference a variable in the manifest YAML file is to use double parentheses `((` and `))`, e.g. `((variableName))`.
 *
 * You can declare variable assignments as key value-pairs inside a YAML variables file following the
 * [Cloud Foundry standards](https://docs.cloudfoundry.org/devguide/deploy-apps/manifest-attributes.html#variable-substitution) format.
 *
 * Optionally, you can also specify a direct list of key-value mappings for variables using the `manifestVariables` parameter.
 * Variables given in the `manifestVariables` list will take precedence over those found in variables files. This follows
 * the behavior of `cf push --var`, and works in combination with `manifestVariablesFiles`.
 *
 * The step is activated by the presence of the file specified by the `manifestFile` parameter and all variables files
 * specified by the `manifestVariablesFiles` parameter, or if variables are passed in directly via `manifestVariables`.
 *
 * In case no `manifestVariablesFiles` were explicitly specified, a default named `manifest-variables.yml` will be looked
 * for and if present will activate this step also. This is to support convention over configuration.
 */
@GenerateDocumentation
void call(Map arguments = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: arguments) {
        def script = checkScript(this, arguments)  ?: this
        String stageName = arguments.stageName ?: env.STAGE_NAME

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(arguments, PARAMETER_KEYS)
            .use()

        String defaultManifestFileName = "manifest.yml"
        String defaultManifestVariablesFileName = "manifest-variables.yml"

        Boolean manifestVariablesFilesExplicitlySpecified = config.manifestVariablesFiles != null

        String manifestFilePath = config.manifestFile ?: defaultManifestFileName
        List<String> manifestVariablesFiles = (config.manifestVariablesFiles != null) ? config.manifestVariablesFiles : [ defaultManifestVariablesFileName ]
        List<Map<String, Object>> manifestVariablesList = config.manifestVariables ?: []
        String outputFilePath = config.outputManifestFile ?: manifestFilePath

        DebugHelper debugHelper = new DebugHelper(script, config)
        YamlUtils yamlUtils = new YamlUtils(script, debugHelper)

        Boolean manifestExists = fileExists manifestFilePath
        Boolean manifestVariablesFilesExist = allManifestVariableFilesExist(manifestVariablesFiles)
        Boolean manifestVariablesListSpecified = !manifestVariablesList.isEmpty()

        if (!manifestExists) {
            echo "[CFManifestSubstituteVariables] Could not find YAML file at ${manifestFilePath}. Skipping variable substitution."
            return
        }

        if (!manifestVariablesFilesExist && manifestVariablesFilesExplicitlySpecified) {
            // If the user explicitly specified a list of variables files, make sure they all exist.
            // Otherwise throw an error so the user knows that he / she made a mistake.
            error "[CFManifestSubstituteVariables] Could not find all given manifest variable substitution files. Make sure all files given as manifestVariablesFiles exist."
        }

        def result
        ExecutionContext context = new ExecutionContext()

        if (!manifestVariablesFilesExist && !manifestVariablesFilesExplicitlySpecified) {
            // If no variables files exist (not even the default one) we check if at least we have a list of variables.

            if (!manifestVariablesListSpecified) {
                // If we have no variable values to replace references with, we skip substitution.
                echo "[CFManifestSubstituteVariables] Could not find any default manifest variable substitution file at ${defaultManifestVariablesFileName}, and no manifest variables list was specified. Skipping variable substitution."
                return
            }

            // If we have a list of variables specified, we can start replacing them...
            result = substitute(manifestFilePath, [], manifestVariablesList, yamlUtils, context, debugHelper)
        }
        else {
            // If we have at least one existing variable substitution file, we can start replacing variables...
            result = substitute(manifestFilePath, manifestVariablesFiles, manifestVariablesList, yamlUtils, context, debugHelper)
        }

        if (!context.variablesReplaced) {
            // If no variables have been replaced at all, we skip writing a file.
            echo "[CFManifestSubstituteVariables] No variables were found or could be replaced in ${manifestFilePath}. Skipping variable substitution."
            return
        }

        // writeYaml won't overwrite the file. You need to delete it first.
        deleteFile(outputFilePath)

        writeYaml file: outputFilePath, data: result

        echo "[CFManifestSubstituteVariables] Replaced variables in ${manifestFilePath}."
        echo "[CFManifestSubstituteVariables] Wrote output file (with variables replaced) at ${outputFilePath}."
    }
}

/*
 * Substitutes variables specified in files and as lists in a given manifest file.
 * @param manifestFilePath - the path to the manifest file to replace variables in.
 * @param manifestVariablesFiles - the paths to variables substitution files.
 * @param manifestVariablesList - the list of variables data to replace variables with.
 * @param yamlUtils - the `YamlUtils` used for variable substitution.
 * @param context - an `ExecutionContext` to examine if any variables have been replaced and should be written.
 * @param debugHelper - a debug output helper.
 * @return an Object graph of Yaml data with variables substituted (if any were found and could be replaced).
 */
private Object substitute(String manifestFilePath, List<String> manifestVariablesFiles, List<Map<String, Object>> manifestVariablesList, YamlUtils yamlUtils, ExecutionContext context, DebugHelper debugHelper) {
    Boolean noVariablesReplaced = true

    def manifestData = loadManifestData(manifestFilePath, debugHelper)

    // replace variables from list first.
    List<Map<String,Object>> reversedManifestVariablesList = manifestVariablesList.reverse() // to make sure last one wins.

    def result = manifestData
    for (Map<String, Object> manifestVariableData : reversedManifestVariablesList) {
        def executionContext = new ExecutionContext()
        result = yamlUtils.substituteVariables(result, manifestVariableData, executionContext)
        noVariablesReplaced = noVariablesReplaced && !executionContext.variablesReplaced // remember if variables were replaced.
    }

    // replace remaining variables from files
    List<String> reversedManifestVariablesFilesList = manifestVariablesFiles.reverse() // to make sure last one wins.
    for (String manifestVariablesFilePath : reversedManifestVariablesFilesList) {
        def manifestVariablesFileData = loadManifestVariableFileData(manifestVariablesFilePath, debugHelper)
        def executionContext = new ExecutionContext()
        result = yamlUtils.substituteVariables(result, manifestVariablesFileData, executionContext)
        noVariablesReplaced = noVariablesReplaced && !executionContext.variablesReplaced // remember if variables were replaced.
    }

    context.variablesReplaced = !noVariablesReplaced
    return result
}

/*
 * Loads the contents of a manifest.yml file by parsing Yaml and returning the
 * object graph. May return a `List<Object>`  (in case more YAML segments are in the file)
 * or a `Map<String, Object>` in case there is just one segment.
 * @param manifestFilePath - the file path of the manifest to parse.
 * @param debugHelper - a debug output helper.
 * @return the parsed object graph.
 */
private Object loadManifestData(String manifestFilePath, DebugHelper debugHelper) {
    try {
        // may return a List<Object>  (in case more YAML segments are in the file)
        // or a Map<String, Object> in case there is just one segment.
        def result = readYaml file: manifestFilePath
        echo "[CFManifestSubstituteVariables] Loaded manifest at ${manifestFilePath}!"
        return result
    }
    catch(Exception ex) {
        debugHelper.debug("Exception: ${ex}")
        echo "[CFManifestSubstituteVariables] Could not load manifest file at ${manifestFilePath}. Exception was: ${ex}"
        throw ex
    }
}

/*
 * Loads the contents of a manifest variables file by parsing Yaml and returning the
 * object graph. May return a `List<Object>`  (in case more YAML segments are in the file)
 * or a `Map<String, Object>` in case there is just one segment.
 * @param variablesFilePath - the path to the variables file to parse.
 * @param debugHelper - a debug output helper.
 * @return the parsed object graph.
 */
private Object loadManifestVariableFileData(String variablesFilePath, DebugHelper debugHelper) {
    try {
        // may return a List<Object>  (in case more YAML segments are in the file)
        // or a Map<String, Object> in case there is just one segment.
        def result = readYaml file: variablesFilePath
        echo "[CFManifestSubstituteVariables] Loaded variables file at ${variablesFilePath}!"
        return result
    }
    catch(Exception ex) {
        debugHelper.debug("Exception: ${ex}")
        echo "[CFManifestSubstituteVariables] Could not load manifest variables file at ${variablesFilePath}. Exception was: ${ex}"
        throw ex
    }
}

/*
 * Checks if all file paths given in the list exist as files.
 * @param manifestVariablesFiles - the list of file paths pointing to manifest variables files.
 * @return `true`, if all given files exist, `false` otherwise.
 */
private boolean allManifestVariableFilesExist(List<String> manifestVariablesFiles) {
    for (String filePath : manifestVariablesFiles) {
        Boolean fileExists = fileExists filePath
        if (!fileExists) {
            echo "[CFManifestSubstituteVariables] Did not find manifest variable substitution file at ${filePath}."
            return false
        }
    }
    return true
}

/*
 * Removes the given file, if it exists.
 * @param filePath - the path to the file to remove.
 */
private void deleteFile(String filePath) {

    Boolean fileExists = fileExists file: filePath
    if(fileExists) {
        Boolean failure = sh script: "rm '${filePath}'", returnStatus: true
        if(!failure) {
            echo "[CFManifestSubstituteVariables] Successfully deleted file '${filePath}'."
        }
        else {
            error "[CFManifestSubstituteVariables] Could not delete file '${filePath}'. Check file permissions."
        }
    }
}
