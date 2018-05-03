package com.sap.piper.tools

import com.sap.piper.VersionUtils
import com.sap.piper.EnvironmentUtils
import com.sap.piper.FileUtils
import com.sap.piper.Version

import hudson.AbortException


class ToolDescriptor implements Serializable {

    final name
    final environmentKey
    final stepConfigurationKey
    final executablePath
    final executableName
    final singleVersion
    final multipleVersions
    final versionOption

    ToolDescriptor(name, environmentKey, stepConfigurationKey, executablePath, executableName, String singleVersion, versionOption) {
        this.name = name
        this.environmentKey = environmentKey
        this.stepConfigurationKey = stepConfigurationKey
        this.executablePath = executablePath
        this.executableName = executableName
        this.singleVersion = singleVersion
        this.multipleVersions = [:]
        this.versionOption = versionOption
    }

    ToolDescriptor(name, environmentKey, stepConfigurationKey, executablePath, executableName, Map multipleVersions, versionOption) {
        this.name = name
        this.environmentKey = environmentKey
        this.stepConfigurationKey = stepConfigurationKey
        this.executablePath = executablePath
        this.executableName = executableName
        this.singleVersion = ''
        this.multipleVersions = multipleVersions
        this.versionOption = versionOption
    }

    def getToolLocation(script, configuration, log = true) {

        def toolLocation
        if (EnvironmentUtils.isEnvironmentVariable(script, environmentKey)) {
            toolLocation = EnvironmentUtils.getEnvironmentVariable(script, environmentKey)
            if (log) script.echo "$name home '$toolLocation' retrieved from environment."
        }
        else if (configuration.containsKey(stepConfigurationKey)) {
            toolLocation = configuration.get(stepConfigurationKey)
            if (log) script.echo "$name home '$toolLocation' retrieved from configuration."
        } else if (isOnPath(script, configuration)){
            toolLocation = ''
            if (log) script.echo "$name is on PATH."
        } else {
            throw new AbortException(getMessage())
        }
        return toolLocation
    }

    def getTool(script, configuration, log = true) {

        def toolLocation = getToolLocation(script, configuration, log)

        if (toolLocation) {
            return "$toolLocation$executablePath$executableName"
        } else {
            return executableName
        }
    }

    def getToolExecutable(script, configuration, log = true) {
        def executable = getTool(script, configuration, log)
        if (log) script.echo "Using $name '$executable'."
        return executable
    }

    def verify(script, configuration) {

        verifyToolLocation(script, configuration)
        verifyToolExecutable(script, configuration)
        verifyVersion(script, configuration)
    }

    def verifyToolLocation(script, configuration) {

        def toolLocation = getToolLocation(script, configuration)
        if (toolLocation) { 
            script.echo "Verifying $name location '$toolLocation'."
            FileUtils.validateDirectoryIsNotEmpty(script, toolLocation)
            script.echo "Verification success. $name location '$toolLocation' exists."
        }
    }

    def verifyToolExecutable(script, configuration) {

        def home = getToolLocation(script, configuration, false)
        def tool = getTool(script, configuration, false)
        if (home) {
            script.echo "Verifying $name '$tool'."
            FileUtils.validateFile(script, tool)
            script.echo "Verification success. $name '$tool' exists."
        }
    }

    def verifyVersion(script, configuration) {

        def executable = getToolExecutable(script, configuration, false)
        if (singleVersion) VersionUtils.verifyVersion(script, name, executable, singleVersion, versionOption)
        if (multipleVersions) VersionUtils.verifyVersion(script, name, executable, multipleVersions, versionOption)
    }

    def getMessage() {
        def configOptions = "Please, configure $name home. $name home can be set "
        if (environmentKey) configOptions += "using the environment variable '$environmentKey', or "
        if (stepConfigurationKey) configOptions += "using the configuration key '$stepConfigurationKey', or "
        configOptions += "on PATH."
        return configOptions
    }

    def isOnPath(script, configuration) {

        def exitStatus
        try {
          exitStatus = script.sh returnStatus: true, script: """set +x
                                                                which $executableName"""
        } catch(AbortException e) {
          throw new AbortException("The verification of $name failed, while checking if it was on PATH. Reason: $e.message.")
        }
        return exitStatus == 0
    }
}
