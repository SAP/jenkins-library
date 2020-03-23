import com.sap.piper.DownloadCacheUtils
import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/mavenExecute.yaml'

void call(Map parameters = [:]) {
    List credentials = [ ]
    parameters = DownloadCacheUtils.injectDownloadCacheInMavenParameters(parameters.script, parameters)
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
