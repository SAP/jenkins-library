package com.sap.piper.tools.neo

import com.sap.piper.BashUtils
import com.sap.piper.Utils

class NeoCommandHelper {

    private Script script
    private String deployMode
    private Map deploymentConfiguration
    private String pathToNeoExecutable
    private String user
    private String password
    private String source

    NeoCommandHelper(Script script, String deployMode, Map deploymentConfiguration, String pathToNeoExecutable,
                    String user, String password, String source) {
        this.script = script
        this.deployMode = deployMode
        this.deploymentConfiguration = deploymentConfiguration
        this.pathToNeoExecutable = pathToNeoExecutable
        this.user = user
        this.password = password
        this.source = source
    }

    private String prolog() {
        return "\"${pathToNeoExecutable}\""
    }

    String statusCommand() {
        if (deployMode == 'mta') {
            throw new Exception("This should not happen. Status command cannot be executed for MTA applications")
        }
        return "${prolog()} status ${mainArgs()}"
    }

    String rollingUpdateCommand() {
        return "${prolog()} rolling-update ${mainArgs()} ${source()} ${additionalArgs()}"
    }

    String deployCommand() {
        "${prolog()} deploy ${mainArgs()} ${source()} ${additionalArgs()}"
    }

    String restartCommand() {
        return "${prolog()} restart --synchronous ${mainArgs()}"
    }

    String deployMta() {
        return "${prolog()} deploy-mta --synchronous ${mainArgs()} ${source()}"
    }

    String cloudCockpitLink() {
        if (deployMode == "warPropertiesFile") {
            Map properties = loadConfigurationFromPropertiesFile()
            return "https://account.${properties.host}/cockpit#" +
                "/acc/${properties.account}/app/${properties.application}/dashboard"
        }

        if (deployMode == "mta") {
            assertMandatoryParameter('host')
            assertMandatoryParameter('account')
            return "https://account.${deploymentConfiguration.host}/cockpit#" +
                "/acc/${deploymentConfiguration.account}/mtas"
        }

        assertMandatoryParameter('host')
        assertMandatoryParameter('account')
        assertMandatoryParameter('application')

        return "https://account.${deploymentConfiguration.host}/cockpit#" +
            "/acc/${deploymentConfiguration.account}/app/${deploymentConfiguration.application}/dashboard"
    }

    String resourceLock() {
        if (deployMode == "warPropertiesFile") {
            Map properties = loadConfigurationFromPropertiesFile()
            return "${properties.host}/${properties.account}/${properties.application}"
        }

        assertMandatoryParameter("host")
        assertMandatoryParameter("account")

        String resource = "${deploymentConfiguration.host}/${deploymentConfiguration.account}"

        if (deployMode == "warParams") {
            assertMandatoryParameter("application")
            resource += "/${deploymentConfiguration.application}"
        }

        return resource
    }

    private String source() {
        assertFileExists(source)
        return "--source ${BashUtils.escape(source)}"
    }

    private String mainArgs() {
        String usernamePassword = "--user ${BashUtils.escape(user)} --password ${BashUtils.escape(password)}"

        if (deployMode == 'warPropertiesFile') {
            assertMandatoryParameter('propertiesFile')
            assertFileIsConfiguredAndExists('propertiesFile')
            return "${deploymentConfiguration.propertiesFile} ${usernamePassword}"
        }

        assertMandatoryParameter('host')
        assertMandatoryParameter('account')

        String targetArgs = "--host ${BashUtils.escape(deploymentConfiguration.host)}"
        targetArgs += " --account ${BashUtils.escape(deploymentConfiguration.account)}"

        if (deployMode == 'warParams') {
            assertMandatoryParameter('application')
            targetArgs += " --application ${BashUtils.escape(deploymentConfiguration.application)}"
        }

        return "${targetArgs} ${usernamePassword}"
    }

    private additionalArgs() {
        if (deployMode != 'warParams') {
            return ""
        }

        String args = ""
        assertMandatoryParameter('runtime')
        args += " --runtime ${BashUtils.escape(deploymentConfiguration.runtime)}"

        assertMandatoryParameter('runtimeVersion')
        args += " --runtime-version ${BashUtils.escape(deploymentConfiguration.runtimeVersion)}"

        if (deploymentConfiguration.size) {
            args += " --size ${BashUtils.escape(deploymentConfiguration.size)}"
        }

        if (deploymentConfiguration.containsKey('environment')) {
            def environment = deploymentConfiguration.environment

            if (!(environment in Map)) {
                script.error("The environment variables for the deployment to Neo have to be defined as a map.");
            }

            def keys = environment.keySet()

            for (int i = 0; i < keys.size(); i++) {
                def key = keys[i]
                def value = environment.get(keys[i])
                args += " --ev ${BashUtils.escape(key)}=${BashUtils.escape(value)}"
            }
        }


        if (deploymentConfiguration.containsKey('vmArguments')) {
            args += " --vm-arguments ${BashUtils.escape(deploymentConfiguration.vmArguments)}"
        }

        return args
    }

    private Map loadConfigurationFromPropertiesFile() {
        assertFileIsConfiguredAndExists('propertiesFile')

        Map properties = script.readProperties file: deploymentConfiguration.propertiesFile
        if (!properties.application || !properties.host || !properties.account) {
            script.error("Error in Neo deployment configuration. Configuration for host, account or application is missing in the properties file.")
        }

        return properties
    }

    private assertFileIsConfiguredAndExists(configurationKey) {
        assertMandatoryParameter(configurationKey)
        assertFileExists(deploymentConfiguration[configurationKey])
    }

    private assertFileExists(filePath) {
        if (!script.fileExists(filePath)) {
            script.error("File ${filePath} cannot be found.")
        }
    }

    private assertMandatoryParameter(configurationKey) {
        if (!deploymentConfiguration[configurationKey]) {
            script.error("Error in Neo deployment configuration. Configuration for ${configurationKey} is missing.")
        }
    }
}
