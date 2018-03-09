package com.sap.piper.tools

import com.sap.piper.EnvironmentUtils


class ToolUtils implements Serializable {

    def static getToolHome(tool, script, configuration, log = true) {

        def home
        if (EnvironmentUtils.isEnvironmentVariable(script, tool.environmentKey)) {
            home = EnvironmentUtils.getEnvironmentVariable(script, tool.environmentKey)
            if (log) script.echo "$tool.name home '$home' retrieved from environment."
        }
        else if (configuration.containsKey(tool.stepConfigurationKey)) {
            home = configuration.get(tool.stepConfigurationKey)
            if (log) script.echo "$tool.name home '$home' retrieved from configuration."
        } else {
            home = ''
            if (log) script.echo "$tool.name expected on PATH or current working directory."
        }
        return home
    }

    def static getToolExecutable(tool, script, configuration, log = true) {

        def home = getToolHome(tool, script, configuration, log)

        def path = "$tool.executablePath"
        def executable = "$tool.executableName"
        def toolExecutable

        if (home) {
            toolExecutable = "$home$path$executable"
        } else {
            toolExecutable = "$executable"
        }
        if (log) script.echo "Using $tool.name executable '$toolExecutable'."
        return toolExecutable
    }
}
