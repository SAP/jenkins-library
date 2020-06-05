package com.sap.piper.variablesubstitution

import hudson.AbortException

/**
 * A utility class for Yaml data.
 * Deals with the substitution of variables within Yaml objects.
 */
class YamlUtils implements Serializable {

    private final DebugHelper logger
    private final Script script

    /**
     * Creates a new utils instance for the given script.
     * @param script - the script which will be used to call pipeline steps.
     * @param logger - an optional debug helper to print debug messages.
     */
    YamlUtils(Script script, DebugHelper logger = null) {
        if(!script) {
            throw new IllegalArgumentException("[YamlUtils] Script must not be null.")
        }
        this.script = script
        this.logger = logger
    }

    /**
     * Substitutes variables references in a given input Yaml object with values that are read from the
     * passed variables Yaml object. Variables may be of primitive or complex types.
     * The format of variable references follows [Cloud Foundry standards](https://docs.cloudfoundry.org/devguide/deploy-apps/manifest-attributes.html#variable-substitution)
     *
     * @param inputYaml - the input Yaml data as `Object`. Can be either of type `Map` or `List`.
     * @param variablesYaml - the variables Yaml data as `Object`. Can be either of type `Map` or `List` and should
     *  contain variables names and values to replace variable references contained in `inputYaml`.
     * @param context - an `ExecutionContext` that can be used to query whether the script actually replaced any variables.
     * @return the YAML object graph of substituted data.
     */
    Object substituteVariables(Object inputYaml, Object variablesYaml, ExecutionContext context = null) {
        if (!inputYaml) {
            throw new IllegalArgumentException("[YamlUtils] Input Yaml data must not be null or empty.")
        }

        if (!variablesYaml) {
            throw new IllegalArgumentException("[YamlUtils] Variables Yaml data must not be null or empty.")
        }

        return substitute(inputYaml, variablesYaml, context)
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
                    logger?.debug("[YamlUtils] WARNING - Found variable reference ${referenceToReplace} in input Yaml but no variable value to replace it with Leaving it unresolved. Check your variables Yaml data and make sure the variable is properly declared.")
                    return manifestNode
                }

                script.echo "[YamlUtils] Replacing: ${referenceToReplace} with ${substitute}"

                if(isSingleVariableReference(stringNode)) {
                    logger?.debug("[YamlUtils] Node ${stringNode} is SINGLE variable reference. Substitute type is: ${substitute.getClass().getName()}")
                    // if the string node we need to do replacements for is
                    // a reference to a single variable, i.e. should be replaced
                    // entirely with the variable value, we replace the entire node
                    // with the variable's value (which can possibly be a complex type).
                    complexResult = substitute
                }
                else {
                    logger?.debug("[YamlUtils] Node ${stringNode} is multi-variable reference or contains additional string constants. Substitute type is: ${substitute.getClass().getName()}")
                    // if the string node we need to do replacements for contains various
                    // variable references or a variable reference and constant string additions
                    // we do a string replacement of the variables inside the node.
                    String regex = "\\(\\(${referenceName}\\)\\)"
                    stringNode = stringNode.replaceAll(regex, substitute as String)
                }
            }

            if (context) {
                context.variablesReplaced = true // remember that variables were found in the YAML file that have been replaced.
            }

            return (complexResult != null) ? complexResult : stringNode
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
            logger?.debug("[YamlUtils] Found data type ${manifestNode.getClass().getName()} that needs no substitute. Value: ${manifestNode}")
            return manifestNode
        }
    }

    /**
     * Turns the parsed variables Yaml data into a
     * single map. Takes care of multiple YAML sections (separated by ---) if they are found and flattens them into a single
     * map if necessary.
     * @param variablesYamlData - the variables data as a Yaml object.
     * @return the `Map` of variable names mapped to their substitute values.
     */
    private Map<String, Object> getVariableSubstitutes(Object variablesYamlData) {

        if(variablesYamlData instanceof List) {
            return flattenVariablesFileData(variablesYamlData as List)
        }
        else if (variablesYamlData instanceof Map) {
            return variablesYamlData as Map<String, Object>
        }
        else {
            // should never happen (hopefully...)
            throw new AbortException("[YamlUtils] Found unsupported data type of variables file after parsing YAML. Expected either List or Map. Got: ${variablesYamlData.getClass().getName()}.")
        }
    }

    /**
     * Flattens a list of Yaml sections (which are deemed to be key-value mappings of variable names and values)
     * to a single map. In case multiple Yaml sections contain the same key, values will be overridden and the result
     * will be undefined.
     * @param variablesYamlData - the `List` of Yaml objects of the different sections.
     * @return the `Map` of variable substitute mappings.
     */
    private Map<String, Object> flattenVariablesFileData(List<Map<String, Object>> variablesYamlData) {
        Map<String, Object> substitutes = new HashMap<>()
        variablesYamlData.each { map ->
            map.entrySet().each { entry ->
                substitutes.put(entry.key, entry.value)
            }
        }
        return substitutes
    }

    /**
     * Returns true, if the given object node contains variable references.
     * @param node - the object-typed value to check for variable references.
     * @return `true`, if this node references at least one variable, `false` otherwise.
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
     * string surrounding, e.g. `This-text-has-((numberOfWords))-words`.
     * @param node - the node to check.
     * @return `true` if the node is a single variable reference. `false` otherwise.
     */
    private boolean isSingleVariableReference(String node) {
        // regex matching only if the entire node is a reference. (^ = matches start of word, $ = matches end of word)
        String regex = '^\\(\\([\\d\\w-]*\\)\\)$' // use single quote not to have to escape $ (interpolation) sign.
        List<String> matches = node.findAll(regex)
        return (matches != null && !matches.isEmpty())
    }

    /**
     * Returns a map of variable references (including braces) to plain variable names referenced in the given `String`.
     * The keys of the map are the variable references, the values are the names of the referenced variables.
     * @param value - the value to look for variable references in.
     * @return the `Map` of names of referenced variables.
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
     * (by stripping braces) of the variable. E.g. input: `((my_var-04))`, output: `my_var-04`
     * @param variableReference - the variable reference including braces.
     * @return the plain variable name
     */
    private String getPlainVariableName(String variableReference) {
        String result = variableReference.replace("((", "")
        result = result.replace("))", "")
        return result
    }
}
