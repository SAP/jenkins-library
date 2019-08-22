package com.sap.piper.variablesubstitution

/**
 * Very simple debug helper. Declared as a Field
 * and passed the configuration with a call to `setConfig(Map config)`
 * in the body of a `call(...)` block, once
 * the configuration is available. <br>
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
     * Creates a new instance.
     */
    DebugHelper() {}

    /**
     * Sets up the debug helper with the given script and config.
     * @param script - the script to use to issue echo statements.
     * @param config - the config whose `verbose` flag will be inspected before echoing messages.
     */
    void setup(Script script, Map config) {
        this.script = script
        this.config = config
    }
    /**
     * log a debug message if a configuration
     * indicates that the `verbose` flag
     * is set to `true`
     * @param message
     */
    void debug(String message) {
        if(script != null && config?.verbose != null) {
            script.echo message
        }
    }
}
