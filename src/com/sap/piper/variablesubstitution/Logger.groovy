package com.sap.piper.variablesubstitution

/**
 * Very simple logger class that can be instantiated to
 * log debug messages. The logger should be declared as
 * a field and a call to {@code setConfig(Map config)} should
 * follow from the body of the {@code call(...)} block, once
 * the configuration is available. <br>
 * If {@code config.verbose} is set to {@code true} the message
 * will be logged. Otherwise it will silently be omitted.
 */
class Logger {
    /**
     * The configuration to check for
     * logging. Make sure to set this
     * configuration with {@code config.verbose}
     * set to {@code true} to log anything.
     */
    Map config

    /**
     * Creates a new instance.
     */
    Logger() {}

    /**
     * log a debug message if a configuration
     * indicates that the {@code verbose} flag
     * is set to {@code true}
     * @param message
     */
    void debug(String message) {
        if(config?.verbose) {
            println(message)
        }
    }
}
