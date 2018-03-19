package com.sap.piper.tools

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
    final version
    final versionOption

    ToolDescriptor(name, environmentKey, stepConfigurationKey, executablePath, executableName, version, versionOption) {
        this.name = name
        this.environmentKey = environmentKey
        this.stepConfigurationKey = stepConfigurationKey
        this.executablePath = executablePath
        this.executableName = executableName
        this.version = version
        this.versionOption = versionOption
    }

    def getHome(script, configuration, log = true) {

        def home
        if (EnvironmentUtils.isEnvironmentVariable(script, environmentKey)) {
            home = EnvironmentUtils.getEnvironmentVariable(script, environmentKey)
            if (log) script.echo "$name home '$home' retrieved from environment."
        }
        else if (configuration.containsKey(stepConfigurationKey)) {
            home = configuration.get(stepConfigurationKey)
            if (log) script.echo "$name home '$home' retrieved from configuration."
        } else {
            home = ''
            if (log) script.echo "$name expected on PATH."
        }
        return home
    }

    def getTool(script, configuration, log = true) {

        def home = getHome(script, configuration, log)

        def path = "$executablePath"
        def executable = "$executableName"

        if (home) {
            return "$home$path$executable"
        } else {
            return "$executable"
        }
    }

    def getExecutable(script, configuration, log = true) {
        def executable = getTool(script, configuration, log)
        if (log) script.echo "Using $name executable '$executable'."
        return executable
    }

    def verify(script, configuration) {

        verifyHome(script, configuration)
        verifyTool(script, configuration)
        verifyVersion(script, configuration)
    }

    def verifyHome(script, configuration) {

        def home = getHome(script, configuration)
        if (home) { 
            script.echo "Verifying $name home '$home'."
            FileUtils.validateDirectoryIsNotEmpty(script, home)
            script.echo "Verification success. $name home '$home' exists."
        }
    }

    def verifyTool(script, configuration) {

        def home = getHome(script, configuration, false)
        def tool = getTool(script, configuration, false)
        if (home) {
            script.echo "Verifying $name '$tool'."
            FileUtils.validateFile(script, tool)
            script.echo "Verification success. $name '$tool' exists."
        }
    }

    def verifyVersion(script, configuration) {

        def executable = getExecutable(script, configuration, false)

        script.echo "Verifying $name version $version or compatible version."

        def toolVersion
        try {
          toolVersion = script.sh returnStdout: true, script: "$executable $versionOption"
        } catch(AbortException e) {
          throw new AbortException("The verification of $name failed. Please check '$executable'. $e.message.")
        }
        def installedVersion = new Version(toolVersion)
        if (!installedVersion.isCompatibleVersion(new Version(version))) {
          throw new AbortException("The installed version of $name is ${installedVersion.toString()}. Please install version $version or a compatible version.")
        }
        script.echo "Verification success. $name version ${installedVersion.toString()} is installed."
    }
}
