package com.sap.piper.variablesubstitution

/**
 * A simple class to capture context information
 * of the execution of yamlSubstituteVariables.
 */
class ExecutionContext {
    /**
     * Property indicating if the execution
     * of yamlSubstituteVariables actually
     * substituted any variables at all.
     *
     * Does NOT indicate that ALL variables were
     * actually replaced. If set to true, if just indicates
     * that some or all variables have been replaced.
     */
    Boolean variablesReplaced = false
}
