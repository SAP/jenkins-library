import com.sap.piper.FileUtils
import com.sap.piper.Version
import com.sap.piper.Utils
import com.sap.piper.ToolUtils
import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger

import hudson.AbortException


 def call(Map parameters = [:]) {

    List parameterKeys = [
        'tool',
        'home'
    ]

    List generalConfigurationKeys = [
        'mtaJarLocation',
        'neoHome',
        'cmCliHome'
    ]

    handlePipelineStepErrors (stepName: 'toolValidate', stepParameters: parameters) {

        final script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]
 
        final Map generalConfiguration = ConfigurationLoader.generalConfiguration(script)
        final Map configuration = ConfigurationMerger.merge(
                                      parameters, parameterKeys,
                                      generalConfiguration, generalConfigurationKeys)
 
        def utils = new Utils()
        def tool = utils.getMandatoryParameter(configuration, 'tool')
        if (configuration.home) FileUtils.validateDirectoryIsNotEmpty(configuration.home)

        switch(tool) {
            case 'java': validateJava(configuration)
            return
            case 'mta': validateMta(configuration)
            return
            case 'neo': validateNeo(configuration)
            return
            case 'cm': validateCm(configuration)
            return
            default:
            throw new AbortException("The tool '$tool' is not supported. The following tools are supported: java, mta, neo and cm.")
        }
    }
}

def validateJava(configuration) {
    def executable = configuration.home ? "$configuration.home/bin/java" : "$JAVA_HOME/bin/java"
    validateTool('Java', executable, "$executable -version 2>&1", new Version(1,8,0))
}

def validateMta(configuration) {
    def executable = configuration.home ? "$configuration.home/mta.jar" : ToolUtils.getMtaJar(this, 'toolValidate', configuration, env)
    validateTool('SAP Multitarget Application Archive Builder', executable, "$JAVA_HOME/bin/java -jar $executable -v", new Version(1, 0, 6))
}

def validateNeo(configuration) {
    def executable = configuration.home ? "$configuration.home/tools/neo.sh" : ToolUtils.getNeoExecutable(this, 'toolValidate', configuration, env)
    validateTool('SAP Cloud Platform Console Client', executable, "$executable version", new Version(3, 39, 10))
}

def validateCm(configuration) {
    def executable = configuration.home ? "$configuration.home/bin/cmclient" : ToolUtils.getCmCliExecutable(this, 'toolValidate', configuration, env)
    validateTool('Change Management Command Line Interface', executable, "$executable -v", new Version(0, 0, 1))
}

private validateTool(name, executable, command, expectedVersion) {
    echo "[toolValidate] Validating $name version ${expectedVersion.toString()} or compatible version."
    def output
    try {
      output = sh returnStdout: true, script: command
    } catch(AbortException e) {
      throw new AbortException("The validation of $name failed. Please check '$executable': $e.message.")
    }
    def version = new Version(output)
    if (!version.isCompatibleVersion(expectedVersion)) {
      throw new AbortException("The installed version of $name is ${version.toString()}. Please install version ${expectedVersion.toString()} or a compatible version.")
    }
    echo "[toolValidate] $name version ${version.toString()} is installed."
}

