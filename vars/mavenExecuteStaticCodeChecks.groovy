import com.sap.piper.DownloadCacheUtils
import groovy.transform.Field

@Field String METADATA_FILE = 'metadata/mavenStaticCodeChecks.yaml'
@Field String STEP_NAME = getClass().getName()

void call(Map parameters = [:]) {
    List credentials = []
    parameters = DownloadCacheUtils.injectDownloadCacheInMavenParameters(parameters.script, parameters)
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
