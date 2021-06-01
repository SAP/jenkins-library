import com.sap.piper.BuildTool
import com.sap.piper.DownloadCacheUtils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String METADATA_FILE = 'metadata/mavenBuild.yaml'
@Field String STEP_NAME = getClass().getName()

void call(Map parameters = [:]) {
    List credentials = [[type: 'token', id: 'altDeploymentRepositoryPasswordId', env: ['PIPER_altDeploymentRepositoryPassword']]]
    final script = checkScript(this, parameters) ?: this
    parameters = DownloadCacheUtils.injectDownloadCacheInParameters(script, parameters, BuildTool.MAVEN)

    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
