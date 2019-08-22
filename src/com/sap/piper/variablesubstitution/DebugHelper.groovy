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
     * Flag to control if the log written by
     * `debug()` should be verbose. If set to false,
     * `debug()` will not log anything.
     */
    Boolean verbose = false

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
        if(verbose) {
            println(message)
        }
    }
}
