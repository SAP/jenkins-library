package com.sap.piper.variablesubstitution

/**
 * Very simple debug helper. Declared as a Field
 * and passed the configuration with a call to `setConfig(Map config)`
 * in the body of a `call(...)` block, once
 * the configuration is available.
 *
 * If `config.verbose` is set to `true` a message
 * issued with `debug(String)` will be logged. Otherwise it will silently be omitted.
 */
class DebugHelper {
    /**
     * The script which will be used to echo debug messages.
     */
    private Script script
    /**
     * The configuration which will be scanned for a `verbose` flag.
     * Only if this is true, will debug messages be written.
     */
    private Map config

    /**
     * Creates a new instance using the given script to issue `echo` commands.
     * The given config's `verbose` flag will decide if a message will be logged or not.
     * @param script - the script whose `echo` command will be used.
     * @param config - the configuration whose `verbose` flag is inspected before logging debug statements.
     */
    DebugHelper(Script script, Map config) {
        if(!script) {
            throw new IllegalArgumentException("[DebugHelper] Script parameter must not be null.")
        }

        if(!config) {
            throw new IllegalArgumentException("[DebugHelper] Config map parameter must not be null.")
        }

        this.script = script
        this.config = config
    }

    /**
     * Logs a debug message if a configuration
     * indicates that the `verbose` flag
     * is set to `true`
     * @param message - the message to log.
     */
    void debug(String message) {
        if(config.verbose) {
            script.echo message
        }
    }
}
