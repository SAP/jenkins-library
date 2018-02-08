import com.sap.piper.FileUtils
import com.sap.piper.Version
import com.sap.piper.Utils
import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger

import hudson.AbortException


def call(Map parameters = [:]) {

    def stepName = 'toolValidate'

    List parameterKeys = [
        'tool',
        'home'
    ]

    List generalConfigurationKeys = [
        'mtaJarLocation',
        'neoHome',
        'cmCliHome'
    ]

    handlePipelineStepErrors (stepName: stepName, stepParameters: parameters) {

        final script = parameters?.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        final Map generalConfiguration = ConfigurationLoader.generalConfiguration(script)
        final Map configuration = ConfigurationMerger.merge(
                                      parameters, parameterKeys,
                                      generalConfiguration, generalConfigurationKeys)

        def utils = new Utils()
        def tool = utils.getMandatoryParameter(configuration, 'tool')
        def home = configuration.home

        if (home) FileUtils.validateDirectoryIsNotEmpty(home)

        switch(tool) {
            case 'java':
                validateJava(home)
                return
            case 'mta':
                if (home) validateMta("$home/mta.jar")
                else validateMta(utils.getMtaJar(this, stepName, configuration, env))
                return
            case 'neo':
                if (home) validateNeo("$home/tools/neo.sh")
                else validateNeo(utils.getNeoExecutable(this, stepName, configuration, env))
                return
            case 'cm':
                if (home) validateCm("$home/bin/cmclient")
                else validateCm(utils.getCmCliExecutable(this, stepName, configuration, env))
                return
            default:
                throw new AbortException("The tool '$tool' is not supported. The following tools are supported: java, mta, neo and cm.")
        }
    }
}

def validateJava(home) {
    validateTool('Java', home, "$home/bin/java -version 2>&1", new Version(1,8,0))
}

def validateMta(executable) {
    FileUtils.validateFile(executable)
    validateTool('SAP Multitarget Application Archive Builder', executable, "$JAVA_HOME/bin/java -jar $executable -v", new Version(1, 0, 6))
}

def validateNeo(executable) {
    FileUtils.validateFile(executable)
    validateTool('SAP Cloud Platform Console Client', executable, "$executable version", new Version(3, 39, 10))
}

def validateCm(executable) {
    FileUtils.validateFile(executable)
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

