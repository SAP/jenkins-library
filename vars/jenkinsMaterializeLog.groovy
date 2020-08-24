import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.k8s.ContainerMap
import hudson.FilePath
import com.sap.piper.Utils
import groovy.transform.Field
import java.util.UUID
import java.io.File
import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.JenkinsUtils

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []

@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS


/**
 * This step allows you to materialize the Jenkins log file of the running build.
 *
 * It acts as a wrapper executing the passed function body.
 *
 * Note: the file that has been created during step execution will be removed automatically.
 */
@GenerateDocumentation
void call(Map parameters = [:], body) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()
        checkScript(this, parameters)
        withMaterializedLogFile(body, jenkinsUtils)
    }
}


@NonCPS
def writeLogToFile(fp) {
    currentBuild.rawBuild.getLogText().writeLogTo(0, fp.write())
}

@NonCPS
def deleteLogFile(fp) {
    if(fp.exists()) {
        fp.delete()
    }
}

def getFilePath(logFileName, jenkinsUtils) {
    def nodeName = env['NODE_NAME']
    if (nodeName == null || nodeName.size() == 0) {
        throw new IllegalArgumentException("Environment variable NODE_NAME is undefined")
    }
    def file = new File(logFileName)
    def instance = jenkinsUtils.getInstance()
    if (instance == null) {
        // fall back
        return new FilePath(file);
    } else {
        def computer = instance.getComputer(nodeName)
        if (computer == null) {
            // fall back
            println "Warning: Jenkins returned computer instance null on node " + nodeName
            return new FilePath(file);
        }
        def channel = computer.getChannel()
        return new FilePath(channel, logFileName)
    }
}


// The method cannot be NonCPS because we call CPS
def withMaterializedLogFile(body, jenkinsUtils) {
    def tempLogFileName = "${env.WORKSPACE}/log-${UUID.randomUUID().toString()}.txt"
    def fp = getFilePath(tempLogFileName, jenkinsUtils)
    writeLogToFile(fp)
    try {
        body(tempLogFileName)
    } finally {
        deleteLogFile(fp)
    }
}
