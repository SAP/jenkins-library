package com.sap.piper.tools.neo

import com.sap.piper.BashUtils
import com.sap.piper.Utils
import com.sap.piper.tools.ToolDescriptor

class NeoCommandHelper {

    private def script
    private Map deploymentConfiguration
    private ToolDescriptor neoToolDescriptor
    private String username
    private String password

    NeoCommandHelper(def script, Map deploymentConfiguration, ToolDescriptor neoToolDescriptor, String username, String password){
        this.script = script
        this.deploymentConfiguration = deploymentConfiguration
        this.neoToolDescriptor = neoToolDescriptor
        this.username = password
        this.password = password
    }

    String neoTool(){
        return neoToolDescriptor.getToolExecutable(this, deploymentConfiguration)
    }

    String statusCommand() {
        if(deploymentConfiguration.deployMode == 'mta'){
            throw new Exception("This should not happen. Status command cannot be executed for MTA applications")
        }
        return "${neoTool()} status ${mainArgs()}"
    }

    String rollingUpdateCommand() {
        return "${neoTool()} rolling-update ${mainArgs()} ${source()} ${additionalArgs()}"
    }

    String deployCommand() {
        "${neoTool()} deploy ${mainArgs()} ${source()} ${additionalArgs()}"
    }

    String restartCommand() {
        return "${neoTool()} restart --synchronous ${mainArgs()}"
    }

    String cloudCockpitLink(){
        if(deploymentConfiguration.deployMode = "warPropertiesFile"){
            Map properties = loadConfigurationFromPropertiesFile()
            return "https://account.${properties.host}/cockpit#" +
                "/acc/${properties.account}/app/${properties.application}/dashboard"
        }

        if(deploymentConfiguration.deployMode = "mta"){
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
        if(deploymentConfiguration.deployMode = "warPropertiesFile"){
            Map properties = loadConfigurationFromPropertiesFile()
            return "${properties.host}/${properties.account}/${properties.application}"
        }

        assertMandatoryParameter("host")
        assertMandatoryParameter("account")

        String resource = "${host}/${account}"

        if(deploymentConfiguration.deployMode = "warParams"){
            assertMandatoryParameter("application")
            resource += "/${application}"
        }

        return resource
    }

    private String source(){
        assertFileIsConfiguredAndExists('source')
        return "--source ${deploymentConfiguration.source}"
    }

    private String mainArgs() {
        String usernamePassword = "-u ${username} -p ${BashUtils.escape(password)}"

        if(deploymentConfiguration.deployMode == 'warPropertiesFile'){
            assertMandatoryParameter('propertiesFile')
            assertFileIsConfiguredAndExists('propertiesFile')
            return "${deploymentConfiguration.propertiesFile} ${usernamePassword}"
        }

        assertMandatoryParameter('host')
        assertMandatoryParameter('account')

        String mainArgs = "--host ${deploymentConfiguration.host} --account ${deploymentConfiguration.account} ${usernamePassword}"

        if(deploymentConfiguration.deployMode == 'warParams'){
            assertMandatoryParameter('application')
            mainArgs += " --application ${deploymentConfiguration.application}"
        }

        return mainArgs
    }

    private additionalArgs() {
        String args = ""

        assertMandatoryParameter('runtime')
        args += " --runtime ${deploymentConfiguration.runtime}"

        assertMandatoryParameter('runtimeVersion')
        args += " --runtime ${deploymentConfiguration.runtimeVersion}"

        if (deploymentConfiguration.size) {
            def sizes = ['lite', 'pro', 'prem', 'prem-plus']
            def size = new Utils().getParameterInValueRange(script, deploymentConfiguration,'size',sizes)

            args += " --size ${size}"
        }

        if (deploymentDescriptor.containsKey('environment')) {
            def environment = deploymentConfiguration.environment

            if(environment !(targetEnvironmentVariables in Map)){
                script.error("The environment variables for the deployment to Neo have to be defined as a map.");
            }

            def keys = environment.keySet()

            for (int i = 0; i < keys.size(); i++) {
                def key = keys[i]
                def value = environment.get(keys[i])
                args += " --ev ${BashUtils.escape(key)}=${BashUtils.escape(value)}"
            }
        }


        if (deploymentDescriptor.containsKey('vmArguments')) {
            args += " --vm-arguments \"${deploymentConfiguration.vmArguments}\""
        }

        return args
    }

    private Map loadConfigurationFromPropertiesFile(){
        assertFileIsConfiguredAndExists('propertiesFile')

        Map properties = script.readProperties file: deploymentConfiguration.propertiesFile
        if(!properties.application || !properties.host || properties.account) {
            script.error("Error in Neo deployment configuration. Configuration for host, account or application is missing in the properties file")
        }

        return properties
    }

    private assertFileIsConfiguredAndExists(configurationKey){
        assertMandatoryParameter(configurationKey)
        if(script.fileExists(deploymentConfiguration[configurationKey])) {
            script.error("File ${deploymentConfiguration[configurationKey]} cannot be found.")
        }
    }

    private assertMandatoryParameter(configurationKey){
        if(!deploymentConfiguration[configurationKey]){
            script.error("Error in Neo deployment configuration. Configuration for ${configurationKey} is missing")
        }
    }
}
