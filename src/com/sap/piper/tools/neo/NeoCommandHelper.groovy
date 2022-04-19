package com.sap.piper.tools.neo

import com.sap.piper.BashUtils
import com.sap.piper.StepAssertions

class NeoCommandHelper {

    private Script step
    private DeployMode deployMode
    private Map deploymentConfiguration
    private Set extensions
    private String user
    private String password
    private String source

    //Warning: Commands generated with this class can contain passwords and should only be used within the step withCredentials
    NeoCommandHelper(Script step, DeployMode deployMode, Map deploymentConfiguration,
                    Set extensions,
                    String user, String password, String source) {
        this.step = step
        this.deployMode = deployMode
        this.deploymentConfiguration = deploymentConfiguration
        this.user = user
        this.password = password
        this.source = source
        this.extensions = extensions ?: []
    }

    private String prolog() {
        return 'neo.sh'
    }

    String statusCommand() {
        if (deployMode == DeployMode.MTA) {
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
        return "${prolog()} deploy-mta --synchronous ${mainArgs()}${extensions()} ${source()}"
    }

    String cloudCockpitLink() {
        if (deployMode == DeployMode.WAR_PROPERTIES_FILE) {
            Map properties = loadConfigurationFromPropertiesFile()
            return "https://account.${properties.host}/cockpit#" +
                "/acc/${properties.account}/app/${properties.application}/dashboard"
        }

        if (deployMode == DeployMode.MTA) {
            return "https://account.${deploymentConfiguration.host}/cockpit#" +
                "/acc/${deploymentConfiguration.account}/mtas"
        }

        return "https://account.${deploymentConfiguration.host}/cockpit#" +
            "/acc/${deploymentConfiguration.account}/app/${deploymentConfiguration.application}/dashboard"
    }

    String resourceLock() {
        if (deployMode == DeployMode.WAR_PROPERTIES_FILE) {
            Map properties = loadConfigurationFromPropertiesFile()
            return "${properties.host}/${properties.account}/${properties.application}"
        }

        String resource = "${deploymentConfiguration.host}/${deploymentConfiguration.account}"

        if (deployMode == DeployMode.WAR_PARAMS) {

            resource += "/${deploymentConfiguration.application}"
        }

        return resource
    }

    private String source() {
        StepAssertions.assertFileExists(step, source)
        return "--source ${BashUtils.quoteAndEscape(source)}"
    }

    private String extensions() {
        if(! this.extensions) return ''
        ' --extensions ' + ((Iterable)this.extensions.collect({ "'${it}'" })).join(',')
    }

    private String mainArgs() {
        String usernamePassword = "--user ${BashUtils.quoteAndEscape(user)} --password ${BashUtils.quoteAndEscape(password)}"

        if (deployMode == DeployMode.WAR_PROPERTIES_FILE) {
            StepAssertions.assertFileIsConfiguredAndExists(step, deploymentConfiguration, 'propertiesFile')
            return "${deploymentConfiguration.propertiesFile} ${usernamePassword}"
        }

        String targetArgs = "--host ${BashUtils.quoteAndEscape(deploymentConfiguration.host)}"
        targetArgs += " --account ${BashUtils.quoteAndEscape(deploymentConfiguration.account)}"

        if (deployMode == DeployMode.WAR_PARAMS) {

            targetArgs += " --application ${BashUtils.quoteAndEscape(deploymentConfiguration.application)}"
        }

        return "${targetArgs} ${usernamePassword}"
    }

    private additionalArgs() {
        if (deployMode != DeployMode.WAR_PARAMS) {
            return ""
        }

        String args = ""
        args += " --runtime ${BashUtils.quoteAndEscape(deploymentConfiguration.runtime)}"
        args += " --runtime-version ${BashUtils.quoteAndEscape(deploymentConfiguration.runtimeVersion)}"

        if (deploymentConfiguration.size) {
            args += " --size ${BashUtils.quoteAndEscape(deploymentConfiguration.size)}"
        }

        if (deploymentConfiguration.containsKey('environment')) {
            def environment = deploymentConfiguration.environment

            if (!(environment in Map)) {
                step.error("The environment variables for the deployment to Neo have to be defined as a map.")
            }

            def keys = environment.keySet()

            for (int i = 0; i < keys.size(); i++) {
                def key = keys[i]
                def value = environment.get(keys[i])
                args += " --ev ${BashUtils.quoteAndEscape(key)}=${BashUtils.quoteAndEscape(value)}"
            }
        }


        if (deploymentConfiguration.containsKey('vmArguments')) {
            args += " --vm-arguments ${BashUtils.quoteAndEscape(deploymentConfiguration.vmArguments)}"
        }
        
        if (deploymentConfiguration.containsKey('azDistribution')) {
            args += " --az-distribution ${BashUtils.quoteAndEscape(deploymentConfiguration.azDistribution)}"
        }

        return args
    }

    private Map loadConfigurationFromPropertiesFile() {
        StepAssertions.assertFileIsConfiguredAndExists(step, deploymentConfiguration, 'propertiesFile')

        Map properties = step.readProperties file: deploymentConfiguration.propertiesFile
        if (!properties.application || !properties.host || !properties.account) {
            step.error("Error in Neo deployment configuration. Configuration for host, account or application is missing in the properties file.")
        }

        return properties
    }
}
