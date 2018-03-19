package com.sap.piper.tools

import com.sap.piper.EnvironmentUtils


class JavaArchiveDescriptor extends ToolDescriptor {

    final javaTool
    final javaOptions

    JavaArchiveDescriptor(name, environmentKey, stepConfigurationKey, executablePath, executableName, version, versionOption, javaTool, javaOptions) {
        super(name, environmentKey, stepConfigurationKey, executablePath, executableName, version, versionOption)
        this.javaTool = javaTool
        this.javaOptions = javaOptions
    }

    @Override
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
            if (log) script.echo "$name expected on current working directory."
        }
        return home
    }

    @Override
    def getExecutable(script, configuration, log = true) {

        def tool = getTool(script, configuration, log)
        def javaExecutable = javaTool.getExecutable(script, configuration, false)
        def executable = "$javaExecutable $javaOptions $tool"
        if (log) script.echo "Using $name executable '$executable'."
        return executable
    }
}
