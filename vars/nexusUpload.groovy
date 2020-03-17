import static com.sap.piper.Prerequisites.checkScript
import static groovy.json.JsonOutput.toJson

import groovy.transform.Field

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/nexusUpload.yaml'

//Metadata maintained in file project://resources/metadata/nexusUpload.yaml

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: null
    if (!script) {
        error "Reference to surrounding pipeline script not provided (script: this)."
    }

    // Make shallow copy of parameters, so we can add/remove (top-level) keys without side-effects to calling code
    parameters = [:] << parameters

    // Backwards compatibility
    if (parameters.credentialsId && !parameters.nexusCredentialsId) {
        parameters.nexusCredentialsId = parameters.credentialsId
    }
    parameters.remove('credentialsId')
    // Remove empty credentials, since the will end up as "net.sf.json.JSONNull"
    // when reading back the config via "piper getConfig --contextConfig" and
    // that in turn will trigger the withCredentials() code-path, but fail to
    // create a binding.
    if (!parameters.nexusCredentialsId) {
        parameters.remove('nexusCredentialsId')
    }

    // Replace 'additionalClassifiers' List with JSON encoded String
    if (parameters.additionalClassifiers) {
        parameters.additionalClassifiers = "${toJson(parameters.additionalClassifiers as List)}"
    }
    // Fall-back to artifactId from configuration if not given
    if (!parameters.artifactId && script.commonPipelineEnvironment.configuration.artifactId) {
        parameters.artifactId = script.commonPipelineEnvironment.configuration.artifactId
    }

    List credentials = [[type: 'usernamePassword', id: 'nexusCredentialsId', env: ['PIPER_username', 'PIPER_password']]]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials, true)
}
