package com.sap.piper.tools


class Tool implements Serializable {

    final name
    final environmentKey
    final stepConfigurationKey
    final executablePath
    final executableName
    final version
    final versionOption

    Tool(name, environmentKey, stepConfigurationKey, executablePath, executableName, version, versionOption) {
        this.name = name
        this.environmentKey = environmentKey
        this.stepConfigurationKey = stepConfigurationKey
        this.executablePath = executablePath
        this.executableName = executableName
        this.version = version
        this.versionOption = versionOption
    }
}
