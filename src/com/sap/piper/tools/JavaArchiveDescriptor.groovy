package com.sap.piper.tools

import com.sap.piper.VersionUtils
import com.sap.piper.EnvironmentUtils
import com.sap.piper.FileUtils

import hudson.AbortException


class JavaArchiveDescriptor implements Serializable {

    final name
    final environmentKey
    final stepConfigurationKey
    final version
    final versionOption
    final javaTool
    final javaOptions

    JavaArchiveDescriptor(name, environmentKey, stepConfigurationKey, version, versionOption, javaTool, javaOptions = '') {
        this.name = name
        this.environmentKey = environmentKey
        this.stepConfigurationKey = stepConfigurationKey
        this.version = version
        this.versionOption = versionOption
        this.javaTool = javaTool
        this.javaOptions = javaOptions
    }

    def getFile(script, configuration, log = true) {

        def javaArchiveFile
        if (EnvironmentUtils.isEnvironmentVariable(script, environmentKey)) {
            javaArchiveFile = EnvironmentUtils.getEnvironmentVariable(script, environmentKey)
            if (log) script.echo "$name file '$javaArchiveFile' retrieved from environment."
            if (!isJavaArchiveFile(javaArchiveFile)) script.error "$environmentKey has an unexpected format."
        }
        else if (configuration.containsKey(stepConfigurationKey)) {
            javaArchiveFile = configuration.get(stepConfigurationKey)
            if (log) script.echo "$name file '$javaArchiveFile' retrieved from configuration."
            if (!isJavaArchiveFile(javaArchiveFile)) script.error "$stepConfigurationKey has an unexpected format."
        } else {
            throw new AbortException(getMessage())
        }
        return javaArchiveFile
    }

    def isJavaArchiveFile(String javaArchiveFile) {
        def group = javaArchiveFile =~ /(.+[\/\\])(\w+[.]jar)/
        if (!group.matches() || group[0].size() == 0) group = javaArchiveFile =~ /(\w+[.]jar)/
        if (!group.matches() || group[0].size() == 0) return false
        return true
    }

    def getCall(script, configuration, log = true) {

        def javaArchiveFile = getFile(script, configuration, log)
        if (log) script.echo "Using $name '$javaArchiveFile'."
        def javaExecutable = javaTool.getToolExecutable(script, configuration, false)
        def javaCall = "$javaExecutable -jar"
        if (javaOptions) javaCall += " $javaOptions"
        return "$javaCall $javaArchiveFile"
    }

    def verify(script, configuration) {

        verifyFile(script, configuration)
        verifyVersion(script, configuration)
    }

    def verifyFile(script, configuration) {

        def javaArchiveFile = getFile(script, configuration, false)
        script.echo "Verifying $name '$javaArchiveFile'."
        FileUtils.validateFile(script, javaArchiveFile)
        script.echo "Verification success. $name '$javaArchiveFile' exists."
    }

    def verifyVersion(script, configuration) {

        def javaArchiveCall = getCall(script, configuration, false)
        VersionUtils.verifyVersion(script, name, javaArchiveCall, version, versionOption)
    }

    def getMessage() {
        def configOptions = "Please, configure $name. $name can be set "
        if (environmentKey) configOptions += "using the environment variable '$environmentKey'"
        if (environmentKey && stepConfigurationKey) configOptions += ", or "
        if (stepConfigurationKey) configOptions += "using the configuration key '$stepConfigurationKey'"
        configOptions += ", or it must be located on the current working directory."
        return configOptions
    }
}
