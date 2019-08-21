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
     * The configuration to check for
     * logging. Make sure to set this
     * configuration with `config.verbose`
     * set to `true` to log anything.
     */
    Map config

    /**
     * Creates a new instance.
     */
    DebugHelper() {}

    /**
     * log a debug message if a configuration
     * indicates that the `verbose` flag
     * is set to `true`
     * @param message
     */
    void debug(String message) {
        if(config?.verbose) {
            println(message)
        }
    }
}
