import com.sap.piper.DownloadCacheUtils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/versioning.yaml'

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this

    List credentials = [
        [type: 'ssh', id: 'gitSshKeyCredentialsId'],
        [type: 'usernamePassword', id: 'gitHttpsCredentialsId', env: ['PIPER_username', 'PIPER_password']],
    ]
    parameters = DownloadCacheUtils.injectDownloadCacheInMavenParameters(script, parameters)
    withEnv(["SSH_KNOWN_HOSTS=${env.JENKINS_HOME}/.ssh/known_hosts"]) {
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
    }
}
