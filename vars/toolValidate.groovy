import com.sap.piper.FileUtils
import com.sap.piper.Version
import com.sap.piper.tools.Tool
import com.sap.piper.tools.ToolVerifier
import hudson.AbortException


def call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: 'toolValidate', stepParameters: parameters) {

        def tool = parameters.tool
        def home = parameters.home

        if (!tool) throw new IllegalArgumentException("The parameter 'tool' can not be null or empty.")
        if (!home) throw new IllegalArgumentException("The parameter 'home' can not be null or empty.")

        FileUtils.validateDirectoryIsNotEmpty(this, home)

        switch(tool) {
            case 'java':
                def java = new Tool('Java', 'JAVA_HOME', '', '/bin/', 'java', '1.8.0', '-version 2>&1')
                ToolVerifier.verifyToolVersion(java, this, [:])
                return
            case 'mta':
                def mta = new Tool('SAP Multitarget Application Archive Builder', 'MTA_JAR_LOCATION', 'mtaJarLocation', '/', 'mta.jar', '1.0.6', '-v')
                ToolVerifier.verifyToolVersion(mta, this, [mtaJarLocation: home])
                return
            case 'neo':
                def neo = new Tool('SAP Cloud Platform Console Client', 'NEO_HOME', 'neoHome', '/tools/', 'neo.sh', '3.39.10', 'version')
                ToolVerifier.verifyToolVersion(neo, this, [neoHome: home])
                return
            case 'cm':
                def cmCli = new Tool('Change Management Command Line Interface', 'CM_CLI_HOME', 'cmCliHome', '/bin/', 'cmclient', '0.0.1', '-v')
                ToolVerifier.verifyToolVersion(cmCli, this, [cmCliHome: home])
                return
            default:
                throw new AbortException("The tool \'$tool\' is not supported. The following tools are supported: java, mta, neo and cm.")
        }
    }
}
