import com.sap.piper.BuildTool
import com.sap.piper.DownloadCacheUtils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/whitesourceExecuteScan.yaml'

//Metadata maintained in file project://resources/metadata/whitesourceExecuteScan.yaml

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    parameters = DownloadCacheUtils.injectDownloadCacheInParameters(script, parameters, BuildTool.MTA)

    List credentials = [
        [type: 'token', id: 'orgAdminUserTokenCredentialsId', env: ['PIPER_orgToken']],
        [type: 'token', id: 'userTokenCredentialsId', env: ['PIPER_userToken']],
        [type: 'token', id: 'githubTokenCredentialsId', env: ['PIPER_githubToken']],
        [type: 'file', id: 'dockerConfigJsonCredentialsId', env: ['PIPER_dockerConfigJSON']],
        [type: 'usernamePassword', id: 'golangPrivateModulesGitTokenCredentialsId', env: ['PIPER_privateModulesGitUsername', 'PIPER_privateModulesGitToken']]
    ]
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
