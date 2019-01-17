package com.sap.piper

class StepAssertions {
    def static assertFileIsConfiguredAndExists(Script script, Map configuration, String configurationKey) {
        assertMandatoryParameter(script, configuration, configurationKey)
        assertFileExists(script, configuration[configurationKey])
    }

    def static  assertFileExists(Script script, String filePath) {
        if (!script.fileExists(filePath)) {
            script.error("File ${filePath} cannot be found.")
        }
    }

    def static assertMandatoryParameter(Script script, Map configuration, String configurationKey) {
        if (!configuration[configurationKey]) {
            script.error("Configuration for ${configurationKey} is missing.")
        }
    }
}
