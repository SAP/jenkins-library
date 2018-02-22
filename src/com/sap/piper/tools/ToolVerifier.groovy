package com.sap.piper.tools

import com.sap.piper.FileUtils
import com.sap.piper.Version
import com.sap.piper.tools.ToolUtils

import hudson.AbortException


class ToolVerifier implements Serializable {

    def static verifyToolHome(tool, script, configuration, environment) {

        def home = ToolUtils.getToolHome(tool, script, configuration, environment)
        if (home) { 
            script.echo "Verifying $tool.name home '$home'."
            FileUtils.validateDirectoryIsNotEmpty(script, home)
            script.echo "Verification success. $tool.name home '$home' exists and it is not empty."
        }
        return home
    }

    def static verifyToolExecutable(tool, script, configuration, environment) {

        def home = verifyToolHome(tool, script, configuration, environment)
        def executable = ToolUtils.getToolExecutable(tool, script, home)
        if (home) {
            script.echo "Verifying $tool.name executable."
            FileUtils.validateFile(executable)
        }
        return executable
    }

    def static verifyToolVersion(tool, script, configuration, environment) {

        def executable = ToolUtils.getToolExecutable(tool, script, configuration, environment)
        if (tool.name == 'SAP Multitarget Application Archive Builder'){
            def java = new Tool('Java', 'JAVA_HOME', '', '/bin/', 'java', '1.8.0', '-version 2>&1')
            def javaExecutable = ToolUtils.getToolExecutable(java, script, configuration, environment)
            executable = "$javaExecutable -jar $executable"
        }

        script.echo "Verifying $tool.name version $tool.version or compatible version."

        def toolVersion
        try {
          toolVersion = script.sh returnStdout: true, script: "$executable $tool.versionOption"
        } catch(AbortException e) {
          throw new AbortException("The verification of $tool.name failed. Please check '$executable'. $e.message.")
        }
        def version = new Version(toolVersion)
        if (!version.isCompatibleVersion(new Version(tool.version))) {
          throw new AbortException("The installed version of $tool.name is ${version.toString()}. Please install version $tool.version or a compatible version.")
        }
        script.echo "$tool.name version ${version.toString()} is installed."
    }
}
