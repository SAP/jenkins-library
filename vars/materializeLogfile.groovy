import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.k8s.ContainerMap
import hudson.FilePath
import com.sap.piper.Utils
import groovy.transform.Field
import java.util.UUID
import com.cloudbees.groovy.cps.NonCPS
import jenkins.model.Jenkins

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []

@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * This step allows you to materialize the Jenkins log file of the running build
 */
@GenerateDocumentation
void call(Map parameters = [:], body) {
	handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
		libraryDocumentationUrl: 'https://github.wdf.sap.corp/pages/ContinuousDelivery/piper-doc/',
		libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
	)
	{
		checkScript(this, parameters) ?: this
		withMaterializedLogFile(body)
	}
}


@NonCPS
def writeLogToFile(logFileName) {
	def channel = Jenkins.get().getComputer(env['NODE_NAME']).getChannel()
	def logInputStream = currentBuild.rawBuild.getLogInputStream()
	def fp = new FilePath(channel, logFileName)

	fp.copyFrom(logInputStream)
	logInputStream.close()
}

@NonCPS
def deleteLogFile(logFileName) {
	def channel = Jenkins.get().getComputer(env['NODE_NAME']).getChannel()
	def logInputStream = currentBuild.rawBuild.getLogInputStream()
	def fp = new FilePath(channel, logFileName)

	if(fp.exists()) {
		fp.delete()
	}
}


// The method cannot be NonCPS because we call CPS
def withMaterializedLogFile(body) {
	def tempLogFileName = "${env.WORKSPACE}/log-${UUID.randomUUID().toString()}.txt"
	writeLogToFile(tempLogFileName)
	try {
		body(tempLogFileName)
	} finally {
		deleteLogFile(tempLogFileName)
	}
}