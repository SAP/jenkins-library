import com.sap.piper.FileUtils
import com.sap.piper.Version
import com.sap.piper.tools.JavaArchiveDescriptor
import com.sap.piper.tools.ToolDescriptor

import hudson.AbortException


def call(Map parameters = [:]) {

    handlePipelineStepErrors (stepName: 'toolValidate', stepParameters: parameters) {

        echo '[WARNING][toolValidate] This step is deprecated, and it will be removed in future versions. Validation is automatically done inside the steps.'

        def tool = parameters.tool
        def home = parameters.home

        if (!tool) throw new IllegalArgumentException("The parameter 'tool' can not be null or empty.")
        if (!home) throw new IllegalArgumentException("The parameter 'home' can not be null or empty.")

        FileUtils.validateDirectoryIsNotEmpty(this, home)

        switch(tool) {
            case 'java':
                def java = new ToolDescriptor('Java', 'JAVA_HOME', '', '/bin/', 'java', '1.8.0', '-version 2>&1')
                java.verifyVersion(this, [:])
                return
            case 'mta':
                def java = new ToolDescriptor('Java', 'JAVA_HOME', '', '/bin/', 'java', '1.8.0', '-version 2>&1')
                def mta = new JavaArchiveDescriptor('SAP Multitarget Application Archive Builder', 'MTA_JAR_LOCATION', 'mtaJarLocation', '1.0.6', '-v', java)
                mta.verifyVersion(this, [mtaJarLocation: home])
                return
            case 'neo':
                def neo = new ToolDescriptor('SAP Cloud Platform Console Client', 'NEO_HOME', 'neoHome', '/tools/', 'neo.sh', null, 'version')
                neo.verifyVersion(this, [neoHome: home])
                return
            case 'cm':
                def cmCli = new ToolDescriptor('Change Management Command Line Interface', 'CM_CLI_HOME', 'cmCliHome', '/bin/', 'cmclient', '0.0.1', '-v')
                cmCli.verifyVersion(this, [cmCliHome: home])
                return
            default:
                throw new AbortException("The tool \'$tool\' is not supported. The following tools are supported: java, mta, neo and cm.")
        }
    }
}
