package com.sap.piper

class StepAssertions {
    def static assertFileIsConfiguredAndExists(Script step, Map configuration, String configurationKey) {
        ConfigurationHelper.newInstance(step, configuration).withMandatoryProperty(configurationKey)
        assertFileExists(step, configuration[configurationKey])
    }

    def static  assertFileExists(Script step, String filePath) {
        if (!step.fileExists(filePath)) {
            step.error("File ${filePath} cannot be found.")
        }
    }
}
