import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.variablesubstitution.ExecutionContext
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

private void debug(String message) {
    // Enable this for debugging.
    // println(message)
}

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS + [
    /**
     * The {@code String} path of the Yaml file to replace variables in. <br>
     * Defaults to "manifest.yml" if not specified otherwise.
     */
    'manifestFile',
    /**
     * The {@code String} path of the Yaml file to produce as output.
     * If not specified this will default to {@code manifestFile} and overwrite it.
     */
    'outputManifestFile',
    /**
     * The {@code String} path of the Yaml file containing the variables' values to use as a replace in the manifest file. <br>
     * Defaults to "manifest-variables.yml" if not specified otherwise
     */
    'variablesFile'
]

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Step that substitutes variables in a given YAML file with those specified in a another. The format to reference a variable
 * in the YAML file is to use double parentheses {@code ((} and {@code ))}, e.g. {@code ((variableName))}.<br>
 * A declaration of a variable and assignment of its value is simply done as a property in the variables YAML file.
 * <p>
 * The format follows <a href="https://docs.cloudfoundry.org/devguide/deploy-apps/manifest-attributes.html#variable-substitution">Cloud Foundry standards</a>.
 * <p>
 * The step is activated by the presence of both a {@code manifest.yml} and a variables file. Names of both files are configurable.
 * <p>
 * Usage: yamlSubstituteVariables manifestFile: "manifest.yml", variablesFile: "manifest-variables.yml"
 *
 * @param arguments - the map of arguments.
 */
@GenerateDocumentation
void call(Map<String, String> arguments) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: arguments) {
        def script = checkScript(this, arguments)  ?: this

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
                                        .loadStepDefaults()
                                        .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
                                        .mixinStageConfig(script.commonPipelineEnvironment, arguments.stageName ?: env.STAGE_NAME, STEP_CONFIG_KEYS)
                                        .mixin(arguments, PARAMETER_KEYS)
                                        .use()

        String manifestFilePath = config.manifestFile ?: "manifest.yml"
        String variablesFilePath = config.variablesFile ?: "manifest-variables.yml"
        String outputFilePath = config.outputManifestFile ?: manifestFilePath

        // Parameter for tests only. No public API.
        // This is to skip manifest file deletion in tests, which
        // would be a nuisance, as it alters the sources repository.
        Boolean skipDeletion = arguments.skipDeletion ?: false

        File manifestFile = new File(manifestFilePath)
        File variablesFile = new File(variablesFilePath)

        if (!manifestFile.exists()) {
            echo "[YamlSubstituteVariables] Could not find YAML file at ${manifestFilePath}. Skipping variable substitution."
            return
        }

        if (!variablesFile.exists()) {
            echo "[YamlSubstituteVariables] Could not find variable substitution file at ${variablesFilePath}. Skipping variable substitution."
            return
        }

        def manifestData = null;
        try {
            // may return a List<Object>  (in case more YAML segments are in the file)
            // or a Map<String, Object> in case there is just one segment.
            manifestData = readYaml file: manifestFilePath
            echo "[YamlSubstituteVariables] Loaded manifest at ${manifestFilePath}!"
        }
        catch(Exception ex) {
            debug("Exception: ${ex}")
            error "[YamlSubstituteVariables] Could not load manifest file at ${manifestFilePath}."
            error "[YamlSubstituteVariables] Exception was: ${ex}"
        }

        def variablesData = null
        try {
            // may return a List<Object>  (in case more YAML segments are in the file)
            // or a Map<String, Object> in case there is just one segment.
            variablesData = readYaml file: variablesFilePath
            echo "[YamlSubstituteVariables] Loaded variables file at ${variablesFilePath}!"
        }
        catch(Exception ex) {
            debug("Exception: ${ex}")
            error "[YamlSubstituteVariables] Could not load manifest variables file at ${variablesFilePath}."
            error "[YamlSubstituteVariables] Exception was: ${ex}"
        }

        // substitute all variables.
        ExecutionContext context = new ExecutionContext()
        def result = substitute(manifestData, variablesData, context)

        if (context.noVariablesReplaced) {
            echo "[YamlSubstituteVariables] No variables were found or could be replaced in ${manifestFilePath}. Skipping variable substitution."
            return
        }

        // writeYaml won't overwrite the file. You need to delete it first.
        deleteOutputFileIfPresent(outputFilePath, skipDeletion)

        writeYaml file: outputFilePath, data: result

        echo "[YamlSubstituteVariables] Replaced variables in ${manifestFilePath} with variables from ${variablesFilePath}."
        echo "[YamlSubstituteVariables] Wrote output file (with variables replaced) at ${outputFilePath}."

        debug("Loaded Manifest: ${manifestData}")
        debug("Loaded Variables: ${variablesData}")
        debug("Result: ${result}")
    }
}

/**
 * Removes the given output file, if it exists.
 * The {@code writeYaml} command won't delete files, so you need to
 * remove the file yourself, in case it already exists.
 * @param outputFilePath the path to the output file to write to.
 */
private void deleteOutputFileIfPresent(String outputFilePath, Boolean skipDeletion) {
    File originalFile = new File(outputFilePath)
    if (originalFile.exists() && !skipDeletion) {
        boolean deleted = originalFile.delete()
        echo "[YamlSubstituteVariables] Deleted output file ${outputFilePath}: ${deleted}"
    }
}

/**
 * Recursively substitutes all variables inside the object tree of the manifest YAML.
 * @param manifestNode - the manifest YAML to replace variables in.
 * @param variablesData - the variables values.
 * @param context - an execution context that can be used to query if any variables were replaced.
 * @return a YAML object graph which has all variables replaced.
 */
private Object substitute(Object manifestNode, Object variablesData, ExecutionContext context) {
    Map<String, Object> variableSubstitutes = getVariableSubstitutes(variablesData)

    if (containsVariableReferences(manifestNode)) {

        Object complexResult = null
        String stringNode = manifestNode as String
        Map<String, String> referencedVariables = getReferencedVariables(stringNode)
        referencedVariables.each { referencedVariable ->
            String referenceToReplace = referencedVariable.getKey()
            String referenceName = referencedVariable.getValue()
            Object substitute = variableSubstitutes.get(referenceName)

            if (null == substitute) {
                echo  "[YamlSubstituteVariables] ERROR - Found variable reference ${referenceToReplace} in manifest but no variable value to replace it with."
                echo  "[YamlSubstituteVariables] ERROR - Leaving it unresolved. Check your manifest variables file and make sure the variable is properly declared."
                echo  "[YamlSubstituteVariables] ERROR - Unresolved variables may lead to follow-up problems (e.g. during a CF deployment). Failing this build."
                error "[YamlSubstituteVariables] Not all variables could be resolved."
            }

            echo "Replacing: ${referenceToReplace} with ${substitute}"

            if(isSingleVariableReference(stringNode)) {
                debug("Node ${stringNode} is SINGLE variable reference. Substitute type is: ${substitute.getClass().getName()}")
                // if the string node we need to do replacements for is
                // a reference to a single variable, i.e. should be replaced
                // entirely with the variable value, we replace the entire node
                // with the variable's value (which can possibly be a complex type).
                complexResult = substitute
            }
            else {
                debug("Node ${stringNode} is multi-variable reference or contains additional string constants. Substitute type is: ${substitute.getClass().getName()}")
                // if the string node we need to do replacements for contains various
                // variable references or a variable reference and constant string additions
                // we do a string replacement of the variables inside the node.
                String regex = "\\(\\(${referenceName}\\)\\)"
                stringNode = stringNode.replaceAll(regex, substitute as String)
            }
        }
        context.noVariablesReplaced = false  // remember that variables were found in the YAML file that have been replaced.
        return complexResult ?: stringNode
    }
    else if (manifestNode instanceof List) {
        List<Object> listNode = manifestNode as List<Object>
        // This copy is only necessary, since Jenkins executes Groovy using
        // CPS (https://wiki.jenkins.io/display/JENKINS/Pipeline+CPS+method+mismatches)
        // and has issues with closures in Java 8 lambda expressions. Otherwise we could replace
        // entries of the list in place (using replaceAll(lambdaExpression))
        List<Object> copy = new ArrayList<>()
        listNode.each { entry ->
            copy.add(substitute(entry, variableSubstitutes, context))
        }
        return copy
    }
    else if(manifestNode instanceof Map) {
        Map<String, Object> mapNode = manifestNode as Map<String, Object>
        // This copy is only necessary to avoid immutability errors reported by Jenkins
        // runtime environment.
        Map<String, Object> copy = new HashMap<>()
        mapNode.entrySet().each { entry ->
            copy.put(entry.getKey(), substitute(entry.getValue(), variableSubstitutes, context))
        }
        return copy
    }
    else {
        debug("[YamlSubstituteVariables] Found data type ${manifestNode.getClass().getName()} that needs no substitute. Value: ${manifestNode}")
        return manifestNode
    }
}
/**
 * Turns the parsed data from a manifest-variables.yml file into a
 * single map. Takes care of multiple YAML sections (separated by ---) if they are found in the manifest-variables.yml.
 * @param variablesFileData - the data parsed frm the manifest-variables.yml file.
 * @return the {@code Map} of variable names mapped to their substitute values.
 */
private Map<String, Object> getVariableSubstitutes(Object variablesFileData) {

    if(variablesFileData instanceof List) {
        return flattenVariablesFileData(variablesFileData as List)
    }
    else if (variablesFileData instanceof Map) {
        return variablesFileData as Map<String, Object>
    }
    else {
        // should never happen (hopefully...)
        error "[YamlSubstituteVariables] Found unsupported data type of variables file after parsing YAML. Expected either List or Map. Got: ${variablesFileData.getClass().getName()}."
    }
}

/**
 * Flattens a list of YAML files (which are deemed to be key-value mappings of variable names and values)
 * to a single map. In case multiple YAML files contain the same key, values will be overridden and the result
 * will be undefined.
 * @param variablesFileData - the {@code List} of YAML objects that have been parsed from a (multi-YAML) manifest-variables.yml.
 * @return the {@code Map} of variable substitute mappings.
 */
private Map<String, Object> flattenVariablesFileData(List<Map<String, Object>> variablesFileData) {
    Map<String, Object> substitutes = new HashMap<>()
    variablesFileData.each { map ->
        map.entrySet().each { entry ->
            substitutes.put(entry.key, entry.value)
        }
    }
    return substitutes
}
/**
 * Returns true, if the given object node contains variable references.
 * @param node the object-typed value to check for variable references.
 * @return {@code true}, if this node references at least one variable, {@code false} otherwise.
 */
private boolean containsVariableReferences(Object node) {
    if(!(node instanceof String)) {
        // variable references can only be contained in
        // string nodes.
        return false
    }
    String stringNode = node as String
    return stringNode.contains("((") && stringNode.contains("))")
}
/**
 * Returns true, if and only if the entire node passed in as a parameter
 * is a variable reference. Returns false if the node references multiple
 * variables or if the node embeds the variable reference inside of a constant
 * string surrounding, e.g. This-text-has-((numberOfWords))-words.
 * @param node the node to check.
 * @return {@code true} if the node is a single variable reference. {@code false} otherwise.
 */
private boolean isSingleVariableReference(String node) {
    // regex matching only if the entire node is a reference. (^ = matches start of word, $ = matches end of word)
    String regex = '^\\(\\([\\d\\w-]*\\)\\)$' // use single quote not to have to escape $ (interpolation) sign.
    List<String> matches = node.findAll(regex)
    return (matches != null && !matches.isEmpty())
}

/**
 * Returns a map of variable references (including braces) to plain variable names referenced in the given {@code String}.
 * The keys of the map are the variable references, the values are the names of the referenced variables.
 * @param value - the value to look for variable references in.
 * @return the {@code Map} of names of referenced variables.
 */
private Map<String, String> getReferencedVariables(String value) {
    Map<String, String> referencesNamesMap = new HashMap<>()
    List<String> variableReferences = value.findAll("\\(\\([\\d\\w-]*\\)\\)") // find all variables in braces, e.g. ((my-var_05))

    variableReferences.each { reference ->
        referencesNamesMap.put(reference, getPlainVariableName(reference))
    }

    return referencesNamesMap
}

/**
 * Expects a variable reference (including braces) as input and returns the plain name
 * (by stripping braces) of the variable. E.g. input: ((my_var-04)) -> output: my_var-04
 * @param variableReference - the variable reference including braces.
 * @return the plain variable name
 */
private String getPlainVariableName(String variableReference) {
    String result = variableReference.replace("((", "")
    result = result.replace("))", "")
    return result
}
