import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/githubrelease.yaml'

void call(Map parameters = [:]) {
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, [[type: 'token', id: 'githubTokenCredentialsId', env: ['PIPER_token']]])
}
