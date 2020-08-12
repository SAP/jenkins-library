import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/protecode.yaml'

//Metadata maintained in file project://resources/metadata/protecode.yaml

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this

    List credentials = [
        [type: 'usernamePassword', id: 'protecodeCredentialsId', env: ['PIPER_username', 'PIPER_password']],
        [type: 'file', id: 'dockerConfigJsonCredentialsId', env: ['DOCKER_CONFIG']],
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
