package com.sap.piper.tools


class ToolUtils implements Serializable {

    def static getToolHome(tool, script, configuration, environment) {

        def home
        if (environment?."$tool.environmentKey") {
            home = environment."$tool.environmentKey"
            script.echo "$tool.name home '$home' retrieved from environment."
        }
        else if (configuration.containsKey(tool.stepConfigurationKey)) {
            home = configuration.get(tool.stepConfigurationKey)
            script.echo "$tool.name home '$home' retrieved from configuration."
        } else {
            home = ''
            script.echo "$tool.name expected on PATH."
        }
        return home
    }

    def static getToolExecutable(tool, script, configuration, environment) {

        def home = getToolHome(tool, script, configuration, environment)
        return getToolExecutable(tool, script, home)
    }

    def static getToolExecutable(tool, script, home) {

        def path = "$tool.executablePath"
        def executable = "$tool.executableName"
        def toolExecutable

        if (home) {
            toolExecutable = "$home$path$executable"
        } else {
            toolExecutable = "$executable"
        }
        script.echo "Using $tool.name executable '$toolExecutable'."
        return toolExecutable
    }
}
